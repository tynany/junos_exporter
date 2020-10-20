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

	totalEnvErrors = 0.0
)

// EnvCollector collects environment metrics, implemented as per the Collector interface.
type EnvCollector struct{}

// NewEnvCollector returns a new EnvCollector.
func NewEnvCollector() *EnvCollector {
	return &EnvCollector{}
}

// Name of the collector.
func (*EnvCollector) Name() string {
	return envSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *EnvCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalEnvErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalEnvErrors
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-environment-information/>`))
	if err != nil {
		totalEnvErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalEnvErrors
	}

	if err := processEnvNetconfReply(reply, ch); err != nil {
		totalEnvErrors++
		errors = append(errors, err)
	}
	return errors, totalEnvErrors
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
