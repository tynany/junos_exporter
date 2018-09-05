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
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9347").String()
	telemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configPath    = kingpin.Flag("config.path", "Path of the YAML configuration file.").Required().String()

	collectors        = []*collector.Collector{}
	exporterSSHConfig = map[string]collector.SSHConfig{}
	collectorConfig   *config.Configuration
)

func initCollectors() {
	iface := collector.NewInterfaceCollector()
	collectors = append(collectors, &collector.Collector{
		PromCollector: iface,
		Errors:        iface,
		Name:          iface.Name(),
	})
	bgp := collector.NewBGPCollector()
	collectors = append(collectors, &collector.Collector{
		PromCollector: bgp,
		Errors:        bgp,
		Name:          bgp.Name(),
	})
	env := collector.NewEnvCollector()
	collectors = append(collectors, &collector.Collector{
		PromCollector: env,
		Errors:        env,
		Name:          env.Name(),
	})
	power := collector.NewPowerCollector()
	collectors = append(collectors, &collector.Collector{
		PromCollector: power,
		Errors:        power,
		Name:          power.Name(),
	})
	re := collector.NewRECollector()
	collectors = append(collectors, &collector.Collector{
		PromCollector: re,
		Errors:        re,
		Name:          re.Name(),
	})
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
	enabledCollectors := []*collector.Collector{}
	for _, collector := range collectors {
		for _, col := range collectorConfig.Config[configParam].Collectors {
			if collector.Name == col {
				enabledCollectors = append(enabledCollectors, collector)
			}
		}
	}

	ne := collector.NewExporter(enabledCollectors)

	if err := ne.SetConnectionDetails(exporterSSHConfig[configParam], targetParam); err != nil {
		log.Errorf("could not set connection details: %s", err)
		return
	}
	registry.Register(ne)

	gatheres := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	}
	promhttp.HandlerFor(gatheres, handlerOpts).ServeHTTP(w, r)
}

func parseCLI() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("junos_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func generateSSHConfig() error {
	for name, configData := range collectorConfig.Config {
		sshConfig := collector.SSHConfig{
			Username: configData.Username,
		}
		if configData.Password != "" {
			sshConfig.Password = configData.Password
		}
		if configData.SSHKey != "" {
			buf, err := ioutil.ReadFile(configData.SSHKey)
			if err != nil {
				return fmt.Errorf("could not open ssh key %q: %s", configData.SSHKey, err)
			}
			sshConfig.SSHKey = buf
		}
		if configData.Timeout != 0 {
			sshConfig.Timeout = time.Second * time.Duration(configData.Timeout)

		} else if collectorConfig.Global.Timeout != 0 {
			sshConfig.Timeout = time.Second * time.Duration(collectorConfig.Global.Timeout)
		} else {
			sshConfig.Timeout = time.Second * 20
		}
		exporterSSHConfig[name] = sshConfig
	}
	return nil
}

func main() {
	prometheus.MustRegister(version.NewCollector("junos_exporter"))

	initCollectors()
	parseCLI()

	log.Infof("Starting junos_exporter %s on %s", version.Info(), *listenAddress)

	var collectorNames []string
	for _, collector := range collectors {
		collectorNames = append(collectorNames, collector.Name)
	}
	var err error
	collectorConfig, err = config.LoadConfigFile(*configPath, collectorNames)
	if err != nil {
		log.Fatal(err)
	}

	if err = generateSSHConfig(); err != nil {
		log.Errorf("could not generate SSH configuration: %s", err)
	}

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
