package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v2"
)

// Configuration contains a slice of all configurations in the format of a map, where the key is the name of the config and the values are a Config type.
type Configuration struct {
	Config map[string]Config `yaml:"configs"`
	Global Global            `yaml:"global"`
}

// Config contains the information required by junos_collector to create SSH based NETCONF connections.
type Config struct {
	Username            string   `yaml:"username"`
	Timeout             int      `yaml:"timeout"`
	Password            string   `yaml:"password"`
	SSHKey              string   `yaml:"ssh_key"`
	AllowedTargets      []string `yaml:"allowed_targets"`
	Collectors          []string `yaml:"enabled_collectors"`
	InterfaceDescKeys   []string `yaml:"interface_description_keys"`
	InterfaceMetricKeys []string `yaml:"interface_metric_keys"`
	BGPTypeKeys         []string `yaml:"bgp_peer_type_keys"`
}

// Global contains the global information required by junos_collector to create SSH based NETCONF connections.
type Global struct {
	AllowedTargets      []string `yaml:"allowed_targets"`
	Timeout             int      `yaml:"timeout"`
	InterfaceDescKeys   []string `yaml:"interface_description_keys"`
	InterfaceMetricKeys []string `yaml:"interface_metric_keys"`
	BGPTypeKeys         []string `yaml:"bgp_peer_type_keys"`
}

// LoadConfigFile returns a Configs type from a passed file.
func LoadConfigFile(path string, collectors []string) (*Configuration, error) {
	var configs Configuration

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %q: %v", path, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&configs); err != nil {
		return nil, fmt.Errorf("could not parse config file %q: %v", path, err)
	}

	if err := parseConfig(&configs, collectors); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &configs, nil
}

func parseConfig(configuration *Configuration, validCollectors []string) error {
	for name, configData := range configuration.Config {
		if configData.Username == "" {
			return fmt.Errorf("missing username in %q configuration", name)
		}

		if configData.Password == "" && configData.SSHKey == "" {
			return fmt.Errorf("missing password or ssh_key in %q configuration", name)
		}

		if len(configData.Collectors) == 0 {
			return fmt.Errorf("no collectors enabled in %q configuration", name)
		}

		if configData.SSHKey != "" {
			buf, err := ioutil.ReadFile(configData.SSHKey)
			if err != nil {
				return fmt.Errorf("could not open ssh_key %q in %q configuration: %v", configData.SSHKey, name, err)
			}
			_, err = ssh.ParsePrivateKey(buf)
			if err != nil {
				return fmt.Errorf("invalid ssh_key %q in %q configuration: %v", configData.SSHKey, name, err)
			}
			for _, collector := range configData.Collectors {
				for _, validCollector := range validCollectors {
					if collector == validCollector {
						goto CollectorFound
					}
				}
				return fmt.Errorf("invalid collector %q in %q configuration", collector, name)
			}
		CollectorFound:
		}
	}
	return nil
}
