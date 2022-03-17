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
		"Status":            colPromDesc(envSubsystem, "module_state", "Module Environmental State (1 = OK, 0 = Not OK).", envLabels),
		"Temp":              colPromDesc(envSubsystem, "module_temperature_celsius", "Module Temperature in Celsius", envLabels),
		"FanNormalSpeed":    colPromDesc(envSubsystem, "module_fan_normal_speed_temperature_celsius", "Fan Normal Speed Temperature Threshold", envLabels),
		"FanHighSpeed":      colPromDesc(envSubsystem, "module_fan_high_speed_temperature_celsius", "Fan High Speed Temperature Threshold", envLabels),
		"BadFanYellowAlarm": colPromDesc(envSubsystem, "module_bad_fan_yellow_alarm_temperature_celsius", "Bad Fan Yellow Alarm Temperature Threshold", envLabels),
		"BadFanRedAlarm":    colPromDesc(envSubsystem, "module_bad_fan_red_alarm_temperature_celsius", "Bad Fan Red Alarm Temperature Threshold", envLabels),
		"YellowAlarm":       colPromDesc(envSubsystem, "module_yellow_alarm_temperature_celsius", "Yellow Alarm Temperature Threshold", envLabels),
		"RedAlarm":          colPromDesc(envSubsystem, "module_red_alarm_temperature_celsius", "Red Alarm Temperature Threshold", envLabels),
		"FireShutdown":      colPromDesc(envSubsystem, "module_fire_shutdown_temperature_celsius", "Fire Shutdown Temperature Threshold", envLabels),
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

	// show chassis environment
	replyEnv, err := s.Exec(netconf.RawMethod(`<get-environment-information/>`))
	if err != nil {
		totalEnvErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalEnvErrors
	}

	// show chassis temperature-threshold
	replyEnvTempThreshold, err := s.Exec(netconf.RawMethod(`<get-temperature-threshold-information/>`))
	if err != nil {
		totalEnvErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalEnvErrors
	}

	if err := processEnvNetconfReply(replyEnv, replyEnvTempThreshold, ch); err != nil {
		totalEnvErrors++
		errors = append(errors, err)
	}

	return errors, totalEnvErrors
}

func processEnvNetconfReply(
	replyEnv *netconf.RPCReply,
	replyEnvTempThreshold *netconf.RPCReply,
	ch chan<- prometheus.Metric,
) error {
	var netconfEnvReply envRPCReply
	var netconfEnvTempThresholdReply envTempThresholdRPCReply

	// ** unmarshal show chassis environment <get-environment-information> START ** //
	if err := xml.Unmarshal([]byte(replyEnv.RawReply), &netconfEnvReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, envData := range netconfEnvReply.EnvInformation.EnvironmentItem {
		labels := []string{strings.TrimSpace(envData.Name.Text)}
		envStatus := 0.0
		if envData.Status.Text == "OK" {
			envStatus = 1.0
		}
		ch <- prometheus.MustNewConstMetric(envDesc["Status"], prometheus.GaugeValue, envStatus, labels...)
		newGauge(ch, envDesc["Temp"], envData.Temperature.Temp, labels...)
	}
	// ** unmarshal show chassis temperature-thresholds <get-environment-information> END ** //

	// ** unmarshal show chassis temperature-thresholds <get-temperature-threshold-information> START ** //
	if err := xml.Unmarshal([]byte(replyEnvTempThreshold.RawReply), &netconfEnvTempThresholdReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}

	for _, envTempThresholdData := range netconfEnvTempThresholdReply.EnvTempThresholdInformation.EnvTempThreshold {
		labels := []string{strings.TrimSpace(envTempThresholdData.Name.Text)}

		newGauge(ch, envDesc["FanNormalSpeed"], envTempThresholdData.FanNormalSpeed.Text, labels...)
		newGauge(ch, envDesc["FanHighSpeed"], envTempThresholdData.FanHighSpeed.Text, labels...)
		newGauge(ch, envDesc["BadFanYellowAlarm"], envTempThresholdData.BadFanYellowAlarm.Text, labels...)
		newGauge(ch, envDesc["BadFanRedAlarm"], envTempThresholdData.BadFanRedAlarm.Text, labels...)
		newGauge(ch, envDesc["YellowAlarm"], envTempThresholdData.YellowAlarm.Text, labels...)
		newGauge(ch, envDesc["RedAlarm"], envTempThresholdData.RedAlarm.Text, labels...)
		newGauge(ch, envDesc["FireShutdown"], envTempThresholdData.FireShutdown.Text, labels...)
	}

	// ** unmarshal show chassis temperature-thresholds <get-temperature-threshold-information> END ** //
	return nil
}

// ********************* show chassis environment START ********************* //
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

// ********************* show chassis environment END ********************* //

// ********************* show chassis temperature-thresholds START ********************* //
type envTempThresholdRPCReply struct {
	EnvTempThresholdInformation EnvTempThresholdInformation `xml:"temperature-threshold-information"`
}

type EnvTempThresholdInformation struct {
	EnvTempThreshold []EnvTempThreshold `xml:"temperature-threshold"`
}

type EnvTempThreshold struct {
	Name              envTempThresholdText `xml:"name"`
	FanNormalSpeed    envTempThresholdText `xml:"fan-normal-speed"`
	FanHighSpeed      envTempThresholdText `xml:"fan-high-speed"`
	BadFanYellowAlarm envTempThresholdText `xml:"bad-fan-yellow-alarm"`
	BadFanRedAlarm    envTempThresholdText `xml:"bad-fan-red-alarm"`
	YellowAlarm       envTempThresholdText `xml:"yellow-alarm"`
	RedAlarm          envTempThresholdText `xml:"red-alarm"`
	FireShutdown      envTempThresholdText `xml:"fire-shutdown"`
}

type envTempThresholdText struct {
	Text string `xml:",chardata"`
}

// ********************* show chassis temperature-thresholds END ********************* //
