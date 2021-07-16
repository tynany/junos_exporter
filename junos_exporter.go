package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/tynany/junos_exporter/collector"
	"github.com/tynany/junos_exporter/config"
	"golang.org/x/crypto/ssh"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9347").String()
	telemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configPath    = kingpin.Flag("config.path", "Path of the YAML configuration file.").Required().String()

	// Slice of all configs.
	collectors = []collector.Collector{}

	// Map of client SSH configuration (value) per config as specified in the config file (key).
	exporterSSHConfig = map[string]*ssh.ClientConfig{}

	// Globally accessible configuration loaded from the config file.
	collectorConfig *config.Configuration

	interfaceDescriptionKeys = map[string][]string{}
	interfaceMetricKeys      = map[string][]string{}
	bgpTypeKeys              = map[string][]string{}
)

func initCollectors() {
	collectors = append(collectors, collector.NewInterfaceCollector())
	collectors = append(collectors, collector.NewBGPCollector())
	collectors = append(collectors, collector.NewEnvCollector())
	collectors = append(collectors, collector.NewPowerCollector())
	collectors = append(collectors, collector.NewRECollector())
  collectors = append(collectors, collector.NewIpsecCollector())
}

func validateRequest(configParam string, targetParam string) error {
	if configParam == "" {
		return fmt.Errorf("'config' parameter must be specified")
	}
	for configName := range collectorConfig.Config {
		if configParam == configName {
			goto ConfigFound
		}
	}
	return fmt.Errorf("could not find %q config in configuration file", configParam)

ConfigFound:
	if targetParam == "" {
		return fmt.Errorf("'target' parameter must be specified")
	}
	if len(collectorConfig.Config[configParam].AllowedTargets) > 0 {
		for _, target := range collectorConfig.Config[configParam].AllowedTargets {
			if targetParam == target {
				goto TargetFound
			}
		}
		return fmt.Errorf("allowed_targets is defined under %q configuration but %q is not listed", configParam, targetParam)
	}
	if len(collectorConfig.Global.AllowedTargets) > 0 {
		for _, target := range collectorConfig.Global.AllowedTargets {
			if targetParam == target {
				goto TargetFound
			}
		}
		return fmt.Errorf("allowed_targets is defined under global configuration but %q is not listed", targetParam)
	}
TargetFound:
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	configParam := r.URL.Query().Get("config")
	targetParam := r.URL.Query().Get("target")
	if err := validateRequest(configParam, targetParam); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	registry := prometheus.NewRegistry()
	enabledCollectors := []collector.Collector{}
	for _, collector := range collectors {
		for _, col := range collectorConfig.Config[configParam].Collectors {
			if collector.Name() == col {
				enabledCollectors = append(enabledCollectors, collector)
			}
		}
	}

	config := collector.Config{
		SSHClientConfig: exporterSSHConfig[configParam],
		SSHTarget:       targetParam,
		IfaceDescrKeys:  interfaceDescriptionKeys[configParam],
		IfaceMetricKeys: interfaceMetricKeys[configParam],
		BGPTypeKeys:     bgpTypeKeys[configParam],
	}

	ne, err := collector.NewExporter(enabledCollectors, config)
	if err != nil {
		log.Errorf("could not start exporter: %s", err)
		return
	}

	registry.Register(ne)

	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	}
	promhttp.HandlerFor(gatherers, handlerOpts).ServeHTTP(w, r)
}

func parseCLI() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("junos_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func generateSSHConfig() error {
	for name, configData := range collectorConfig.Config {
		sshClientConfig := &ssh.ClientConfig{
			User: configData.Username,
		}
		if configData.SSHKey != "" {
			buf, err := ioutil.ReadFile(configData.SSHKey)
			if err != nil {
				return fmt.Errorf("could not open ssh key %q: %s", configData.SSHKey, err)
			}
			parsedKey, err := ssh.ParsePrivateKey(buf)
			if err != nil {
				return err
			}
			sshClientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(parsedKey)}
		} else {
			sshClientConfig.Auth = []ssh.AuthMethod{ssh.Password(configData.Password)}
		}
		if configData.Timeout != 0 {
			sshClientConfig.Timeout = time.Second * time.Duration(configData.Timeout)

		} else if collectorConfig.Global.Timeout != 0 {
			sshClientConfig.Timeout = time.Second * time.Duration(collectorConfig.Global.Timeout)
		} else {
			sshClientConfig.Timeout = time.Second * 20
		}
		sshClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		exporterSSHConfig[name] = sshClientConfig
	}
	return nil
}

func getInterfaceDescriptionKeys() {
	var globalIfaceDesc []string
	if len(interfaceDescriptionKeys) == 0 {
		if len(collectorConfig.Global.InterfaceDescKeys) > 0 {
			for _, descrKey := range collectorConfig.Global.InterfaceDescKeys {
				globalIfaceDesc = append(globalIfaceDesc, descrKey)
			}
		}
	}
	for name, configData := range collectorConfig.Config {
		if len(configData.InterfaceDescKeys) > 0 {
			for _, descrKey := range configData.InterfaceDescKeys {
				interfaceDescriptionKeys[name] = append(interfaceDescriptionKeys[name], descrKey)
			}
		} else {
			interfaceDescriptionKeys[name] = globalIfaceDesc
		}
	}
}

func getInterfaceMetricKeys() {
	var globalIfaceMetrics []string
	if len(interfaceMetricKeys) == 0 {
		if len(collectorConfig.Global.InterfaceMetricKeys) > 0 {
			for _, descrKey := range collectorConfig.Global.InterfaceMetricKeys {
				globalIfaceMetrics = append(globalIfaceMetrics, descrKey)
			}
		}
	}
	for name, configData := range collectorConfig.Config {
		if len(configData.InterfaceMetricKeys) > 0 {
			for _, metricKey := range configData.InterfaceMetricKeys {
				interfaceMetricKeys[name] = append(interfaceMetricKeys[name], metricKey)
			}
		} else {
			interfaceMetricKeys[name] = globalIfaceMetrics
		}
	}
}

func getBGPTypeKeys() {
	var globalBGPTypeKeys []string
	if len(bgpTypeKeys) == 0 {
		if len(collectorConfig.Global.BGPTypeKeys) > 0 {
			for _, descrKey := range collectorConfig.Global.BGPTypeKeys {
				globalBGPTypeKeys = append(globalBGPTypeKeys, descrKey)
			}
		}
	}
	for name, configData := range collectorConfig.Config {
		if len(configData.BGPTypeKeys) > 0 {
			for _, metricKey := range configData.BGPTypeKeys {
				bgpTypeKeys[name] = append(bgpTypeKeys[name], metricKey)
			}
		} else {
			bgpTypeKeys[name] = globalBGPTypeKeys
		}
	}
}

func main() {
	prometheus.MustRegister(version.NewCollector("junos_exporter"))

	initCollectors()
	parseCLI()

	log.Infof("Starting junos_exporter %s on %s", version.Info(), *listenAddress)

	// Get a list of collector names to validate collectors specified in the config file exist.
	var collectorNames []string
	for _, collector := range collectors {
		collectorNames = append(collectorNames, collector.Name())
	}
	var err error
	collectorConfig, err = config.LoadConfigFile(*configPath, collectorNames)
	if err != nil {
		log.Fatal(err)
	}

	if err = generateSSHConfig(); err != nil {
		log.Errorf("could not generate SSH configuration: %s", err)
	}

	getInterfaceDescriptionKeys()
	getInterfaceMetricKeys()
	getBGPTypeKeys()

	http.HandleFunc(*telemetryPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>junos Exporter</title></head>
			<body>
			<h1>junos Exporter</h1>
			<p><a href="` + *telemetryPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}
}
