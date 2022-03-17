package collector

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	opticsSubsystem = "optics"

	opticsLabel = []string{"interface"}
	opticsDesc  = map[string]*prometheus.Desc{
		"ModuleTemperature":                   colPromDesc(opticsSubsystem, "module_temperature", "Module Temperature", opticsLabel),
		"ModuleVoltage":                       colPromDesc(opticsSubsystem, "module_voltage", "Module Voltage", opticsLabel),
		"ModuleTemperatureHighAlarm":          colPromDesc(opticsSubsystem, "module_temperature_high_alarm", "Module Temperature High Alarm", opticsLabel),
		"ModuleTemperatureLowAlarm":           colPromDesc(opticsSubsystem, "module_temperature_low_alarm", "Module Temperature Low Alarm", opticsLabel),
		"ModuleTemperatureHighWarn":           colPromDesc(opticsSubsystem, "module_temperature_high_warn", "Module Temperature High Warn", opticsLabel),
		"ModuleTemperatureLowWarn":            colPromDesc(opticsSubsystem, "module_temperature_low_warn", "Module Temperature Low Warn", opticsLabel),
		"ModuleVoltageHighAlarm":              colPromDesc(opticsSubsystem, "module_voltage_high_alarm", "Module Voltage High Alarm", opticsLabel),
		"ModuleVoltageLowAlarm":               colPromDesc(opticsSubsystem, "module_voltage_low_alarm", "Module Voltage Low Alarm", opticsLabel),
		"ModuleVoltageHighWarn":               colPromDesc(opticsSubsystem, "module_voltage_high_warn", "Module Voltage High Warn", opticsLabel),
		"ModuleVoltageLowWarn":                colPromDesc(opticsSubsystem, "module_voltage_low_warn", "Module Voltage Low Warn", opticsLabel),
		"ModuleTemperatureHighAlarmThreshold": colPromDesc(opticsSubsystem, "module_temperature_high_alarm_threshold", "Module Temperature High Alarm Threshold", opticsLabel),
		"ModuleTemperatureLowAlarmThreshold":  colPromDesc(opticsSubsystem, "module_temperature_low_alarm_threshold", "Module Temperature Low Alarm Threshold", opticsLabel),
		"ModuleTemperatureHighWarnThreshold":  colPromDesc(opticsSubsystem, "module_temperature_high_warn_threshold", "Module Temperature High Warn Threshold", opticsLabel),
		"ModuleTemperatureLowWarnThreshold":   colPromDesc(opticsSubsystem, "module_temperature_low_warn_threshold", "Module Temperature Low Warn Threshold", opticsLabel),
		"ModuleVoltageHighAlarmThreshold":     colPromDesc(opticsSubsystem, "module_voltage_high_alarm_threshold", "Module Voltage High Alarm Threshold", opticsLabel),
		"ModuleVoltageLowAlarmThreshold":      colPromDesc(opticsSubsystem, "module_voltage_low_alarm_threshold", "Module Voltage Low Alarm Threshold", opticsLabel),
		"ModuleVoltageHighWarnThreshold":      colPromDesc(opticsSubsystem, "module_voltage_high_warn_threshold", "Module Voltage High Warn Threshold", opticsLabel),
		"ModuleVoltageLowWarnThreshold":       colPromDesc(opticsSubsystem, "module_voltage_low_warn_threshold", "Module Voltage Low Warn Threshold", opticsLabel),
		"LaserBiasCurrentHighAlarmThreshold":  colPromDesc(opticsSubsystem, "laser_bias_current_high_alarm_threshold", "Laser Bias Current High Alarm Threshold", opticsLabel),
		"LaserBiasCurrentLowAlarmThreshold":   colPromDesc(opticsSubsystem, "laser_bias_current_low_alarm_threshold", "Laser Bias Current Low Alarm Threshold", opticsLabel),
		"LaserBiasCurrentHighWarnThreshold":   colPromDesc(opticsSubsystem, "laser_bias_current_high_warn_threshold", "Laser Bias Current High Warn Threshold", opticsLabel),
		"LaserBiasCurrentLowWarnThreshold":    colPromDesc(opticsSubsystem, "laser_bias_current_low_warn_threshold", "Laser Bias Current Low Warn Threshold", opticsLabel),
		"LaserTxPowerHighAlarmThreshold":      colPromDesc(opticsSubsystem, "laser_tx_power_high_alarm_threshold", "Laser Tx Power High Alarm Threshold", opticsLabel),
		"LaserTxPowerHighAlarmThresholdDbm":   colPromDesc(opticsSubsystem, "laser_tx_power_high_alarm_threshold_dbm", "Laser Tx Power High Alarm Threshold Dbm", opticsLabel),
		"LaserTxPowerLowAlarmThreshold":       colPromDesc(opticsSubsystem, "laser_tx_power_low_alarm_threshold", "Laser Tx Power Low Alarm Threshold", opticsLabel),
		"LaserTxPowerLowAlarmThresholdDbm":    colPromDesc(opticsSubsystem, "laser_tx_power_low_alarm_threshold_dbm", "Laser Tx Power Low Alarm Threshold Dbm", opticsLabel),
		"LaserTxPowerHighWarnThreshold":       colPromDesc(opticsSubsystem, "laser_tx_power_high_warn_threshold", "Laser Tx Power High Warn Threshold", opticsLabel),
		"LaserTxPowerHighWarnThresholdDbm":    colPromDesc(opticsSubsystem, "laser_tx_power_high_warn_threshold_dbm", "Laser Tx Power High Warn Threshold Dbm", opticsLabel),
		"LaserTxPowerLowWarnThreshold":        colPromDesc(opticsSubsystem, "laser_tx_power_low_warn_threshold", "Laser Tx Power Low Warn Threshold", opticsLabel),
		"LaserTxPowerLowWarnThresholdDbm":     colPromDesc(opticsSubsystem, "laser_tx_power_low_warn_threshold_dbm", "Laser Tx Power Low Warn Threshold Dbm", opticsLabel),
		"LaserRxPowerHighAlarmThreshold":      colPromDesc(opticsSubsystem, "laser_rx_power_high_alarm_threshold", "Laser Rx Power High Alarm Threshold", opticsLabel),
		"LaserRxPowerHighAlarmThresholdDbm":   colPromDesc(opticsSubsystem, "laser_rx_power_high_alarm_threshold_dbm", "Laser Rx Power High Alarm Threshold Dbm", opticsLabel),
		"LaserRxPowerLowAlarmThreshold":       colPromDesc(opticsSubsystem, "laser_rx_power_low_alarm_threshold", "Laser Rx Power Low Alarm Threshold", opticsLabel),
		"LaserRxPowerLowAlarmThresholdDbm":    colPromDesc(opticsSubsystem, "laser_rx_power_low_alarm_threshold_dbm", "Laser Rx Power Low Alarm Threshold Dbm", opticsLabel),
		"LaserRxPowerHighWarnThreshold":       colPromDesc(opticsSubsystem, "laser_rx_power_high_warn_threshold", "Laser Rx Power High Warn Threshold", opticsLabel),
		"LaserRxPowerHighWarnThresholdDbm":    colPromDesc(opticsSubsystem, "laser_rx_power_high_warn_threshold_dbm", "Laser Rx Power High Warn Threshold Dbm", opticsLabel),
		"LaserRxPowerLowWarnThreshold":        colPromDesc(opticsSubsystem, "laser_rx_power_low_warn_threshold", "Laser Rx Power Low Warn Threshold", opticsLabel),
		"LaserRxPowerLowWarnThresholdDbm":     colPromDesc(opticsSubsystem, "laser_rx_power_low_warn_threshold_dbm", "Laser Rx Power Low Warn Threshold Dbm", opticsLabel),
		"LaneIndex":                           colPromDesc(opticsSubsystem, "lane_index", "Lane Index", opticsLabel),
		"LaserBiasCurrent":                    colPromDesc(opticsSubsystem, "laser_bias_current", "Laser Bias Current", opticsLabel),
		"LaserOutputPower":                    colPromDesc(opticsSubsystem, "laser_output_power", "Laser Output Power", opticsLabel),
		"LaserOutputPowerDbm":                 colPromDesc(opticsSubsystem, "laser_output_power_dbm", "Laser Output Power Dbm", opticsLabel),
		"LaserRxOpticalPower":                 colPromDesc(opticsSubsystem, "laser_rx_optical_power", "Laser Rx Optical Power", opticsLabel),
		"LaserRxOpticalPowerDbm":              colPromDesc(opticsSubsystem, "laser_rx_optical_power_dbm", "Laser Rx Optical Power Dbm", opticsLabel),
		"LaserBiasCurrentHighAlarm":           colPromDesc(opticsSubsystem, "laser_bias_current_high_alarm", "Laser Bias Current High Alarm", opticsLabel),
		"LaserBiasCurrentLowAlarm":            colPromDesc(opticsSubsystem, "laser_bias_current_low_alarm", "Laser Bias Current Low Alarm", opticsLabel),
		"LaserBiasCurrentHighWarn":            colPromDesc(opticsSubsystem, "laser_bias_current_high_warn", "Laser Bias Current High Warn", opticsLabel),
		"LaserBiasCurrentLowWarn":             colPromDesc(opticsSubsystem, "laser_bias_current_low_warn", "Laser Bias Current Low Warn", opticsLabel),
		"LaserRxPowerHighAlarm":               colPromDesc(opticsSubsystem, "laser_rx_power_high_alarm", "Laser Rx Power High Alarm", opticsLabel),
		"LaserRxPowerLowAlarm":                colPromDesc(opticsSubsystem, "laser_rx_power_low_alarm", "Laser Rx Power Low Alarm", opticsLabel),
		"LaserRxPowerHighWarn":                colPromDesc(opticsSubsystem, "laser_rx_power_high_warn", "Laser Rx Power High Warn", opticsLabel),
		"LaserRxPowerLowWarn":                 colPromDesc(opticsSubsystem, "laser_rx_power_low_warn", "Laser Rx Power Low Warn", opticsLabel),
		"TxLossOfSignalFunctionalityAlarm":    colPromDesc(opticsSubsystem, "tx_loss_of_signal_functionality_alarm", "Tx Loss Of Signal Functionality Alarm", opticsLabel),
		"RxLossOfSignalAlarm":                 colPromDesc(opticsSubsystem, "rx_loss_of_signal_alarm", "Rx Loss Of Signal Alarm", opticsLabel),
		"TxLaserDisabledAlarm":                colPromDesc(opticsSubsystem, "tx_laser_disabled_alarm", "Tx Laser Disabled Alarm", opticsLabel),
	}

	totalOpticsErrors = 0.0
)

// OpticsCollector collects power metrics, implemented as per the Collector interface.
type OpticsCollector struct{}

// NewOpticsCollector returns a new OpticsCollector .
func NewOpticsCollector() *OpticsCollector {
	return &OpticsCollector{}
}

// Name of the collector.
func (*OpticsCollector) Name() string {
	return opticsSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *OpticsCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalOpticsErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalOpticsErrors
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-interface-optics-diagnostics-information></get-interface-optics-diagnostics-information>`))
	if err != nil {
		totalOpticsErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalOpticsErrors
	}

	if err := processOpticsNetconfReply(reply, ch); err != nil {
		totalOpticsErrors++
		errors = append(errors, err)
	}
	return errors, totalOpticsErrors
}

func processOpticsNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
	var netconfReply opticsRPCReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, opticsData := range netconfReply.OpticsInterfaceInformation.OpticsPhysicalInterface {
		labels := []string{strings.TrimSpace(opticsData.Name.Text)}

	}
	/*
			labels := []string{strings.TrimSpace(powerData.Name.Text)}

			powerState := 0.0
			if powerData.State.Text == "Online" {
				powerState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(powerDesc["State"], prometheus.GaugeValue, powerState, labels...)
			newGauge(ch, powerDesc["CapacityActual"], powerData.PemCapacityDetail.CapacityActual.Text, labels...)
			newGauge(ch, powerDesc["CapacityMax"], powerData.PemCapacityDetail.CapacityMax.Text, labels...)

			powerACInputState := 0.0
			if powerData.AcInputDetail.AcInput.Text == "OK" {
				powerACInputState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(powerDesc["ACInputState"], prometheus.GaugeValue, powerACInputState, labels...)
			newGauge(ch, powerDesc["ACExpectedFeeds"], powerData.AcInputDetail.AcExpectFeed.Text, labels...)
			newGauge(ch, powerDesc["ACConnectedFeeds"], powerData.AcInputDetail.AcActualFeed.Text, labels...)

			labelsDCOutput := []string{strings.TrimSpace(powerData.Name.Text), strings.TrimSpace(powerData.DcOutputDetail.Zone.Text)}
			newGauge(ch, powerDesc["DCOutputOptics"], powerData.DcOutputDetail.DcOptics.Text, labelsDCOutput...)
			newGauge(ch, powerDesc["DCOutputCurrent"], powerData.DcOutputDetail.DcCurrent.Text, labelsDCOutput...)
			newGauge(ch, powerDesc["DCOutputVoltage"], powerData.DcOutputDetail.DcVoltage.Text, labelsDCOutput...)
			newGauge(ch, powerDesc["DCOutputLoad"], powerData.DcOutputDetail.DcLoad.Text, labelsDCOutput...)
		}
		for _, powerSystem := range netconfReply.OpticsUsageInformation.OpticsUsageSystem {
			for _, zone := range powerSystem.OpticsUsageZoneInformation {
				labels := []string{strings.TrimSpace(zone.Zone.Text)}
				newGauge(ch, powerDesc["CapacityZoneActual"], zone.CapacityActual.Text, labels...)
				newGauge(ch, powerDesc["CapacityZoneMax"], zone.CapacityMax.Text, labels...)
				newGauge(ch, powerDesc["CapacityZoneAllocated"], zone.CapacityAllocated.Text, labels...)
				newGauge(ch, powerDesc["CapacityZoneRemaining"], zone.CapacityRemaining.Text, labels...)
				newGauge(ch, powerDesc["CapacityZoneUsage"], zone.CapacityActualUsage.Text, labels...)
			}
			newGauge(ch, powerDesc["CapacitySysActual"], powerSystem.CapacitySysActual.Text)
			newGauge(ch, powerDesc["CapacitySysMax"], powerSystem.CapacitySysMax.Text)
			newGauge(ch, powerDesc["CapacitySysRemaining"], powerSystem.CapacitySysRemaining.Text)
		}
		for _, fruItem := range netconfReply.OpticsUsageInformation.OpticsUsageFruItem {
			newGauge(ch, powerDesc["DCUsage"], fruItem.DcOptics.Text, fruItem.Name.Text)
		}
	*/
	return nil
}

type opticsRPCReply struct {
	OpticsInterfaceInformation opticsInterfaceInformation `xml:"interface-information"`
}
type opticsInterfaceInformation struct {
	OpticsPhysicalInterface []opticsPhysicalInterface `xml:"physical-interface"`
}

type opticsPhysicalInterface struct {
	Name              opticsText        `xml:"name"`
	OpticsDiagnostics OpticsDiagnostics `xml:"optics-diagnostics"`
}

type OpticsDiagnostics struct {
	ModuleTemperature                   opticsTemp        `xml:"module-temperature"`
	ModuleVoltage                       opticsText        `xml:"module-voltage"`
	ModuleTemperatureHighAlarm          opticsText        `xml:"module-temperature-high-alarm"`
	ModuleTemperatureLowAlarm           opticsText        `xml:"module-temperature-low-alarm"`
	ModuleTemperatureHighWarn           opticsText        `xml:"module-temperature-high-warn"`
	ModuleTemperatureLowWarn            opticsText        `xml:"module-temperature-low-warn"`
	ModuleVoltageHighAlarm              opticsText        `xml:"module-voltage-high-alarm"`
	ModuleVoltageLowAlarm               opticsText        `xml:"module-voltage-low-alarm"`
	ModuleVoltageHighWarn               opticsText        `xml:"module-voltage-high-warn"`
	ModuleVoltageLowWarn                opticsText        `xml:"module-voltage-low-warn"`
	ModuleTemperatureHighAlarmThreshold opticsTemp        `xml:"module-temperature-high-alarm-threshold"`
	ModuleTemperatureLowAlarmThreshold  opticsTemp        `xml:"module-temperature-low-alarm-threshold"`
	ModuleTemperatureHighWarnThreshold  opticsTemp        `xml:"module-temperature-high-warn-threshold"`
	ModuleTemperatureLowWarnThreshold   opticsTemp        `xml:"module-temperature-low-warn-threshold"`
	ModuleVoltageHighAlarmThreshold     opticsText        `xml:"module-voltage-high-alarm-threshold"`
	ModuleVoltageLowAlarmThreshold      opticsText        `xml:"module-voltage-low-alarm-threshold"`
	ModuleVoltageHighWarnThreshold      opticsText        `xml:"module-voltage-high-warn-threshold"`
	ModuleVoltageLowWarnThreshold       opticsText        `xml:"module-voltage-low-warn-threshold"`
	LaserBiasCurrentHighAlarmThreshold  opticsText        `xml:"laser-bias-current-high-alarm-threshold"`
	LaserBiasCurrentLowAlarmThreshold   opticsText        `xml:"laser-bias-current-low-alarm-threshold"`
	LaserBiasCurrentHighWarnThreshold   opticsText        `xml:"laser-bias-current-high-warn-threshold"`
	LaserBiasCurrentLowWarnThreshold    opticsText        `xml:"laser-bias-current-low-warn-threshold"`
	LaserTxPowerHighAlarmThreshold      opticsText        `xml:"laser-tx-power-high-alarm-threshold"`
	LaserTxPowerHighAlarmThresholdDbm   opticsText        `xml:"laser-tx-power-high-alarm-threshold-dbm"`
	LaserTxPowerLowAlarmThreshold       opticsText        `xml:"laser-tx-power-low-alarm-threshold"`
	LaserTxPowerLowAlarmThresholdDbm    opticsText        `xml:"laser-tx-power-low-alarm-threshold-dbm"`
	LaserTxPowerHighWarnThreshold       opticsText        `xml:"laser-tx-power-high-warn-threshold"`
	LaserTxPowerHighWarnThresholdDbm    opticsText        `xml:"laser-tx-power-high-warn-threshold-dbm"`
	LaserTxPowerLowWarnThreshold        opticsText        `xml:"laser-tx-power-low-warn-threshold"`
	LaserTxPowerLowWarnThresholdDbm     opticsText        `xml:"laser-tx-power-low-warn-threshold-dbm"`
	LaserRxPowerHighAlarmThreshold      opticsText        `xml:"laser-rx-power-high-alarm-threshold"`
	LaserRxPowerHighAlarmThresholdDbm   opticsText        `xml:"laser-rx-power-high-alarm-threshold-dbm"`
	LaserRxPowerLowAlarmThreshold       opticsText        `xml:"laser-rx-power-low-alarm-threshold"`
	LaserRxPowerLowAlarmThresholdDbm    opticsText        `xml:"laser-rx-power-low-alarm-threshold-dbm"`
	LaserRxPowerHighWarnThreshold       opticsText        `xml:"laser-rx-power-high-warn-threshold"`
	LaserRxPowerHighWarnThresholdDbm    opticsText        `xml:"laser-rx-power-high-warn-threshold-dbm"`
	LaserRxPowerLowWarnThreshold        opticsText        `xml:"laser-rx-power-low-warn-threshold"`
	LaserRxPowerLowWarnThresholdDbm     opticsText        `xml:"laser-rx-power-low-warn-threshold-dbm"`
	OpticsDiagLanes                     []OpticsDiagLanes `xml:"optics-diagnostics-lane-values"`
}

type OpticsDiagLanes struct {
	LaneIndex                        opticsText `xml:"lane-index"`
	LaserBiasCurrent                 opticsText `xml:"laser-bias-current"`
	LaserOutputPower                 opticsText `xml:"laser-output-power"`
	LaserOutputPowerDbm              opticsText `xml:"laser-output-power-dbm"`
	LaserRxOpticalPower              opticsText `xml:"laser-rx-optical-power"`
	LaserRxOpticalPowerDbm           opticsText `xml:"laser-rx-optical-power-dbm"`
	LaserBiasCurrentHighAlarm        opticsText `xml:"laser-bias-current-high-alarm"`
	LaserBiasCurrentLowAlarm         opticsText `xml:"laser-bias-current-low-alarm"`
	LaserBiasCurrentHighWarn         opticsText `xml:"laser-bias-current-high-warn"`
	LaserBiasCurrentLowWarn          opticsText `xml:"laser-bias-current-low-warn"`
	LaserRxPowerHighAlarm            opticsText `xml:"laser-rx-power-high-alarm"`
	LaserRxPowerLowAlarm             opticsText `xml:"laser-rx-power-low-alarm"`
	LaserRxPowerHighWarn             opticsText `xml:"laser-rx-power-high-warn"`
	LaserRxPowerLowWarn              opticsText `xml:"laser-rx-power-low-warn"`
	TxLossOfSignalFunctionalityAlarm opticsText `xml:"tx-loss-of-signal-functionality-alarm"`
	RxLossOfSignalAlarm              opticsText `xml:"rx-loss-of-signal-alarm"`
	TxLaserDisabledAlarm             opticsText `xml:"tx-laser-disabled-alarm"`
}

type opticsText struct {
	Text string `xml:",chardata"`
}

type opticsTemp struct {
	Temp string `xml:"celsius,attr"`
}
