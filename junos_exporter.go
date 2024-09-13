package main

import (
	"fmt"
	inbuiltLog "log"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/tynany/junos_exporter/collector"
	"github.com/tynany/junos_exporter/config"
	"golang.org/x/crypto/ssh"
)

var (
	telemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configPath    = kingpin.Flag("config.path", "Path of the YAML configuration file.").Required().String()
	webFlagConfig = kingpinflag.AddFlags(kingpin.CommandLine, ":9347")

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

func initCollectors(logger log.Logger) {
	collectors = append(collectors, collector.NewInterfaceCollector(logger))
	collectors = append(collectors, collector.NewBGPCollector(logger))
	collectors = append(collectors, collector.NewEnvCollector(logger))
	collectors = append(collectors, collector.NewPowerCollector(logger))
	collectors = append(collectors, collector.NewRECollector(logger))
	collectors = append(collectors, collector.NewIpsecCollector(logger))
	collectors = append(collectors, collector.NewOpticsCollector(logger))
	collectors = append(collectors, collector.NewOSPFCollector(logger))
	collectors = append(collectors, collector.NewFPCCollector(logger))
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

func handler(logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		nc, err := collector.NewExporter(enabledCollectors, config, logger)
		if err != nil {
			level.Error(logger).Log("msg", "could not create collector", "err", err)
			os.Exit(1)
		}

		if err := registry.Register(nc); err != nil {
			level.Error(logger).Log("msg", "could not register collector", "err", err)
			os.Exit(1)
		}

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		handlerOpts := promhttp.HandlerOpts{
			ErrorLog:      inbuiltLog.New(log.NewStdlibAdapter(level.Error(logger)), "", 0),
			ErrorHandling: promhttp.ContinueOnError,
		}

		metricsHandler := promhttp.HandlerFor(gatherers, handlerOpts)
		metricsHandler.ServeHTTP(w, r)
	})
}

func generateSSHConfig() error {
	for name, configData := range collectorConfig.Config {
		sshClientConfig := &ssh.ClientConfig{
			User: configData.Username,
		}
		if configData.SSHKey != "" {
			buf, err := os.ReadFile(configData.SSHKey)
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
			globalIfaceDesc = append(globalIfaceDesc, collectorConfig.Global.InterfaceDescKeys...)
		}
	}
	for name, configData := range collectorConfig.Config {
		if len(configData.InterfaceDescKeys) > 0 {
			interfaceDescriptionKeys[name] = append(interfaceDescriptionKeys[name], configData.InterfaceDescKeys...)
		} else {
			interfaceDescriptionKeys[name] = globalIfaceDesc
		}
	}
}

func getInterfaceMetricKeys() {
	var globalIfaceMetrics []string
	if len(interfaceMetricKeys) == 0 {
		if len(collectorConfig.Global.InterfaceMetricKeys) > 0 {
			globalIfaceMetrics = append(globalIfaceMetrics, collectorConfig.Global.InterfaceMetricKeys...)
		}
	}
	for name, configData := range collectorConfig.Config {
		if len(configData.InterfaceMetricKeys) > 0 {
			interfaceMetricKeys[name] = append(interfaceMetricKeys[name], configData.InterfaceMetricKeys...)
		} else {
			interfaceMetricKeys[name] = globalIfaceMetrics
		}
	}
}

func getBGPTypeKeys() {
	var globalBGPTypeKeys []string
	if len(bgpTypeKeys) == 0 {
		if len(collectorConfig.Global.BGPTypeKeys) > 0 {
			globalBGPTypeKeys = append(globalBGPTypeKeys, collectorConfig.Global.BGPTypeKeys...)
		}
	}
	for name, configData := range collectorConfig.Config {
		if len(configData.BGPTypeKeys) > 0 {
			bgpTypeKeys[name] = append(bgpTypeKeys[name], configData.BGPTypeKeys...)
		} else {
			bgpTypeKeys[name] = globalBGPTypeKeys
		}
	}
}

func main() {
	promlogConfig := &promlog.Config{}

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("junos_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	initCollectors(logger)

	prometheus.MustRegister(versioncollector.NewCollector("junos_exporter"))

	level.Info(logger).Log("msg", "Starting junos_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	// Get a list of collector names to validate collectors specified in the config file exist.
	var collectorNames []string
	for _, collector := range collectors {
		collectorNames = append(collectorNames, collector.Name())
	}
	var err error
	collectorConfig, err = config.LoadConfigFile(*configPath, collectorNames)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	if err = generateSSHConfig(); err != nil {
		level.Error(logger).Log("could not generate SSH configuration", err)
	}

	getInterfaceDescriptionKeys()
	getInterfaceMetricKeys()
	getBGPTypeKeys()

	http.Handle(*telemetryPath, handler(logger))
	if *telemetryPath != "/" && *telemetryPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Junos Exporter",
			Description: "Prometheus Exporter for Junos",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{Address: *telemetryPath, Text: "Metrics"},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, webFlagConfig, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
