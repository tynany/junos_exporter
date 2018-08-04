package collector

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	envSubsystem = "environment"

	envLabels = []string{"module"}
	envDesc   = map[string]*prometheus.Desc{
		"Status": colPromDesc(envSubsystem, "module_state", "Module Environmental State (1 = OK, 0 = Not OK).", envLabels),
		"Temp":   colPromDesc(envSubsystem, "module_temperature_celsius", "Module Temperature in Celsius", envLabels),
	}

	envErrors      = []error{}
	totalEnvErrors = 0.0
)

// EnvCollector collects environment metrics, implemented as per the Collector env.
type EnvCollector struct{}

// NewEnvCollector returns a EnvCollector type.
func NewEnvCollector() *EnvCollector {
	return &EnvCollector{}
}

// Name of the collector. Used to parse the configuration file.
func (*EnvCollector) Name() string {
	return envSubsystem
}

// CollectErrors returns what errors have been gathered.
func (*EnvCollector) CollectErrors() []error {
	errors := envErrors
	envErrors = []error{}
	return errors
}

// CollectTotalErrors collects total errors.
func (*EnvCollector) CollectTotalErrors() float64 {
	return totalEnvErrors
}

// Describe all metrics implemented as per the prometheus.Collector interface.
func (*EnvCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range envDesc {
		ch <- desc
	}
}

// Collect metrics as per the prometheus.Collector interface.
func (c *EnvCollector) Collect(ch chan<- prometheus.Metric) {
	s, err := netconf.DialSSH(sshTarget, sshClientConfig)
	if err != nil {
		totalEnvErrors++
		envErrors = append(envErrors, fmt.Errorf("could not connect to %q: %s", sshTarget, err))
		return
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-environment-information/>`))
	if err != nil {
		totalEnvErrors++
		envErrors = append(envErrors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return
	}

	if err := processEnvNetconfReply(reply, ch); err != nil {
		totalEnvErrors++
		envErrors = append(envErrors, err)
	}
}

func processEnvNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
	var netconfReply envRPCReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, envData := range netconfReply.EnvInformation.EnvironmentItem {
		labels := []string{strings.TrimSpace(envData.Name.Text)}
		envStatus := 0.0
		if envData.Status.Text == "OK" {
			envStatus = 1.0
		}
		ch <- prometheus.MustNewConstMetric(envDesc["Status"], prometheus.GaugeValue, envStatus, labels...)
		newGauge(ch, envDesc["Temp"], envData.Temperature.Temp, labels...)
	}
	return nil
}

type envRPCReply struct {
	EnvInformation envInformation `xml:"environment-information"`
}

type envInformation struct {
	EnvironmentItem []envItem `xml:"environment-item"`
}

type envItem struct {
	Name envText `xml:"name"`
	// Class       envText `xml:"class"`
	Status      envText `xml:"status"`
	Temperature envTemp `xml:"temperature"`
	// Comment     envText `xml:"comment"`
}

type envText struct {
	Text string `xml:",chardata"`
}

type envTemp struct {
	Temp string `xml:"celsius,attr"`
}
