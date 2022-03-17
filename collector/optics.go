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

	opticsLabel     = []string{"interface"}
	opticsLaneLabel = []string{"interface", "lane"}
	opticsDesc      = map[string]*prometheus.Desc{
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
		"LaneIndex":                           colPromDesc(opticsSubsystem, "lane_index", "Lane Index", opticsLaneLabel),
		"LaserBiasCurrent":                    colPromDesc(opticsSubsystem, "laser_bias_current", "Laser Bias Current", opticsLaneLabel),
		"LaserOutputPower":                    colPromDesc(opticsSubsystem, "laser_output_power", "Laser Output Power", opticsLaneLabel),
		"LaserOutputPowerDbm":                 colPromDesc(opticsSubsystem, "laser_output_power_dbm", "Laser Output Power Dbm", opticsLaneLabel),
		"LaserRxOpticalPower":                 colPromDesc(opticsSubsystem, "laser_rx_optical_power", "Laser Rx Optical Power", opticsLaneLabel),
		"LaserRxOpticalPowerDbm":              colPromDesc(opticsSubsystem, "laser_rx_optical_power_dbm", "Laser Rx Optical Power Dbm", opticsLaneLabel),
		"LaserBiasCurrentHighAlarm":           colPromDesc(opticsSubsystem, "laser_bias_current_high_alarm", "Laser Bias Current High Alarm", opticsLaneLabel),
		"LaserBiasCurrentLowAlarm":            colPromDesc(opticsSubsystem, "laser_bias_current_low_alarm", "Laser Bias Current Low Alarm", opticsLaneLabel),
		"LaserBiasCurrentHighWarn":            colPromDesc(opticsSubsystem, "laser_bias_current_high_warn", "Laser Bias Current High Warn", opticsLaneLabel),
		"LaserBiasCurrentLowWarn":             colPromDesc(opticsSubsystem, "laser_bias_current_low_warn", "Laser Bias Current Low Warn", opticsLaneLabel),
		"LaserRxPowerHighAlarm":               colPromDesc(opticsSubsystem, "laser_rx_power_high_alarm", "Laser Rx Power High Alarm", opticsLaneLabel),
		"LaserRxPowerLowAlarm":                colPromDesc(opticsSubsystem, "laser_rx_power_low_alarm", "Laser Rx Power Low Alarm", opticsLaneLabel),
		"LaserRxPowerHighWarn":                colPromDesc(opticsSubsystem, "laser_rx_power_high_warn", "Laser Rx Power High Warn", opticsLaneLabel),
		"LaserRxPowerLowWarn":                 colPromDesc(opticsSubsystem, "laser_rx_power_low_warn", "Laser Rx Power Low Warn", opticsLaneLabel),
		"TxLossOfSignalFunctionalityAlarm":    colPromDesc(opticsSubsystem, "tx_loss_of_signal_functionality_alarm", "Tx Loss Of Signal Functionality Alarm", opticsLaneLabel),
		"RxLossOfSignalAlarm":                 colPromDesc(opticsSubsystem, "rx_loss_of_signal_alarm", "Rx Loss Of Signal Alarm", opticsLaneLabel),
		"TxLaserDisabledAlarm":                colPromDesc(opticsSubsystem, "tx_laser_disabled_alarm", "Tx Laser Disabled Alarm", opticsLaneLabel),
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

		newGauge(ch, opticsDesc["ModuleTemperature"], opticsData.OpticsDiagnostics.ModuleTemperature.Temp, labels...)
		newGauge(ch, opticsDesc["ModuleVoltage"], opticsData.OpticsDiagnostics.ModuleVoltage.Text, labels...)

		opticsTempHighAlarm := 0.0
		if opticsData.OpticsDiagnostics.ModuleTemperatureHighAlarm.Text == "off" {
			opticsTempHighAlarm = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleTemperatureHighAlarm"], prometheus.GaugeValue, opticsTempHighAlarm, labels...)

		opticsTempLowAlarm := 0.0
		if opticsData.OpticsDiagnostics.ModuleTemperatureLowAlarm.Text == "off" {
			opticsTempHighAlarm = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleTemperatureLowAlarm"], prometheus.GaugeValue, opticsTempLowAlarm, labels...)

		opticsTempHighWarn := 0.0
		if opticsData.OpticsDiagnostics.ModuleTemperatureHighWarn.Text == "off" {
			opticsTempHighWarn = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleTemperatureHighWarn"], prometheus.GaugeValue, opticsTempHighWarn, labels...)

		opticsTempLowWarn := 0.0
		if opticsData.OpticsDiagnostics.ModuleTemperatureLowWarn.Text == "off" {
			opticsTempHighWarn = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleTemperatureLowWarn"], prometheus.GaugeValue, opticsTempLowWarn, labels...)

		opticsVoltageHighAlarm := 0.0
		if opticsData.OpticsDiagnostics.ModuleVoltageHighAlarm.Text == "off" {
			opticsVoltageHighAlarm = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleVoltageHighAlarm"], prometheus.GaugeValue, opticsVoltageHighAlarm, labels...)

		opticsVoltageLowAlarm := 0.0
		if opticsData.OpticsDiagnostics.ModuleVoltageLowAlarm.Text == "off" {
			opticsVoltageHighAlarm = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleVoltageLowAlarm"], prometheus.GaugeValue, opticsVoltageLowAlarm, labels...)

		opticsVoltageHighWarn := 0.0
		if opticsData.OpticsDiagnostics.ModuleVoltageHighWarn.Text == "off" {
			opticsVoltageHighWarn = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleVoltageHighWarn"], prometheus.GaugeValue, opticsVoltageHighWarn, labels...)

		opticsVoltageLowWarn := 0.0
		if opticsData.OpticsDiagnostics.ModuleVoltageLowWarn.Text == "off" {
			opticsVoltageHighWarn = 1.0
		}
		ch <- prometheus.MustNewConstMetric(opticsDesc["ModuleVoltageLowWarn"], prometheus.GaugeValue, opticsVoltageLowWarn, labels...)

		newGauge(ch, opticsDesc["ModuleTemperatureHighAlarmThreshold"], opticsData.OpticsDiagnostics.ModuleTemperatureHighAlarmThreshold.Temp, labels...)
		newGauge(ch, opticsDesc["ModuleTemperatureLowAlarmThreshold"], opticsData.OpticsDiagnostics.ModuleTemperatureLowAlarmThreshold.Temp, labels...)
		newGauge(ch, opticsDesc["ModuleTemperatureHighWarnThreshold"], opticsData.OpticsDiagnostics.ModuleTemperatureHighWarnThreshold.Temp, labels...)
		newGauge(ch, opticsDesc["ModuleTemperatureLowWarnThreshold"], opticsData.OpticsDiagnostics.ModuleTemperatureLowWarnThreshold.Temp, labels...)
		newGauge(ch, opticsDesc["ModuleVoltageHighAlarmThreshold"], opticsData.OpticsDiagnostics.ModuleVoltageHighAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["ModuleVoltageLowAlarmThreshold"], opticsData.OpticsDiagnostics.ModuleVoltageLowAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["ModuleVoltageHighWarnThreshold"], opticsData.OpticsDiagnostics.ModuleVoltageHighWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["ModuleVoltageLowWarnThreshold"], opticsData.OpticsDiagnostics.ModuleVoltageLowWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserBiasCurrentHighAlarmThreshold"], opticsData.OpticsDiagnostics.LaserBiasCurrentHighAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserBiasCurrentLowAlarmThreshold"], opticsData.OpticsDiagnostics.LaserBiasCurrentLowAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserBiasCurrentHighWarnThreshold"], opticsData.OpticsDiagnostics.LaserBiasCurrentHighWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserBiasCurrentLowWarnThreshold"], opticsData.OpticsDiagnostics.LaserBiasCurrentLowWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerHighAlarmThreshold"], opticsData.OpticsDiagnostics.LaserTxPowerHighAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerHighAlarmThresholdDbm"], opticsData.OpticsDiagnostics.LaserTxPowerHighAlarmThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerLowAlarmThreshold"], opticsData.OpticsDiagnostics.LaserTxPowerLowAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerLowAlarmThresholdDbm"], opticsData.OpticsDiagnostics.LaserTxPowerLowAlarmThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerHighWarnThreshold"], opticsData.OpticsDiagnostics.LaserTxPowerHighWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerHighWarnThresholdDbm"], opticsData.OpticsDiagnostics.LaserTxPowerHighWarnThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerLowWarnThreshold"], opticsData.OpticsDiagnostics.LaserTxPowerLowWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserTxPowerLowWarnThresholdDbm"], opticsData.OpticsDiagnostics.LaserTxPowerLowWarnThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerHighAlarmThreshold"], opticsData.OpticsDiagnostics.LaserRxPowerHighAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerHighAlarmThresholdDbm"], opticsData.OpticsDiagnostics.LaserRxPowerHighAlarmThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerLowAlarmThreshold"], opticsData.OpticsDiagnostics.LaserRxPowerLowAlarmThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerLowAlarmThresholdDbm"], opticsData.OpticsDiagnostics.LaserRxPowerLowAlarmThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerHighWarnThreshold"], opticsData.OpticsDiagnostics.LaserRxPowerHighWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerHighWarnThresholdDbm"], opticsData.OpticsDiagnostics.LaserRxPowerHighWarnThresholdDbm.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerLowWarnThreshold"], opticsData.OpticsDiagnostics.LaserRxPowerLowWarnThreshold.Text, labels...)
		newGauge(ch, opticsDesc["LaserRxPowerLowWarnThresholdDbm"], opticsData.OpticsDiagnostics.LaserRxPowerLowWarnThresholdDbm.Text, labels...)

		for _, lane := range opticsData.OpticsDiagnostics.OpticsDiagLanes {
			laneIndex := strings.TrimSpace(lane.LaneIndex.Text)
			laneLabels := append(labels, laneIndex)
			newGauge(ch, opticsDesc["LaneIndex"], lane.LaneIndex.Text, laneLabels...)
			newGauge(ch, opticsDesc["LaserBiasCurrent"], lane.LaserBiasCurrent.Text, laneLabels...)
			newGauge(ch, opticsDesc["LaserOutputPower"], lane.LaserOutputPower.Text, laneLabels...)
			newGauge(ch, opticsDesc["LaserOutputPowerDbm"], lane.LaserOutputPowerDbm.Text, laneLabels...)
			newGauge(ch, opticsDesc["LaserRxOpticalPower"], lane.LaserRxOpticalPower.Text, laneLabels...)
			newGauge(ch, opticsDesc["LaserRxOpticalPowerDbm"], lane.LaserRxOpticalPowerDbm.Text, laneLabels...)

			opticLaneLaserBiasCurrentHighAlarm := 0.0
			if lane.LaserBiasCurrentHighAlarm.Text == "off" {
				opticLaneLaserBiasCurrentHighAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserBiasCurrentHighAlarm"], prometheus.GaugeValue, opticLaneLaserBiasCurrentHighAlarm, laneLabels...)

			opticLaneLaserBiasCurrentLowAlarm := 0.0
			if lane.LaserBiasCurrentLowAlarm.Text == "off" {
				opticLaneLaserBiasCurrentLowAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserBiasCurrentLowAlarm"], prometheus.GaugeValue, opticLaneLaserBiasCurrentLowAlarm, laneLabels...)

			opticLaneLaserBiasCurrentHighWarn := 0.0
			if lane.LaserBiasCurrentHighWarn.Text == "off" {
				opticLaneLaserBiasCurrentHighWarn = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserBiasCurrentHighWarn"], prometheus.GaugeValue, opticLaneLaserBiasCurrentHighWarn, laneLabels...)

			opticLaneLaserBiasCurrentLowWarn := 0.0
			if lane.LaserBiasCurrentLowWarn.Text == "off" {
				opticLaneLaserBiasCurrentLowWarn = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserBiasCurrentLowWarn"], prometheus.GaugeValue, opticLaneLaserBiasCurrentLowWarn, laneLabels...)

			opticLaneLaserRxPowerHighAlarm := 0.0
			if lane.LaserRxPowerHighAlarm.Text == "off" {
				opticLaneLaserRxPowerHighAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserRxPowerHighAlarm"], prometheus.GaugeValue, opticLaneLaserRxPowerHighAlarm, laneLabels...)

			opticLaneLaserRxPowerLowAlarm := 0.0
			if lane.LaserRxPowerLowAlarm.Text == "off" {
				opticLaneLaserRxPowerLowAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserRxPowerLowAlarm"], prometheus.GaugeValue, opticLaneLaserRxPowerLowAlarm, laneLabels...)

			opticLaneLaserRxPowerHighWarn := 0.0
			if lane.LaserRxPowerHighWarn.Text == "off" {
				opticLaneLaserRxPowerHighWarn = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserRxPowerHighWarn"], prometheus.GaugeValue, opticLaneLaserRxPowerHighWarn, laneLabels...)

			opticLaneLaserRxPowerLowWarn := 0.0
			if lane.LaserRxPowerLowWarn.Text == "off" {
				opticLaneLaserRxPowerLowWarn = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["LaserRxPowerLowWarn"], prometheus.GaugeValue, opticLaneLaserRxPowerLowWarn, laneLabels...)

			opticLaneTxLossOfSignalFunctionalityAlarm := 0.0
			if lane.TxLossOfSignalFunctionalityAlarm.Text == "off" {
				opticLaneTxLossOfSignalFunctionalityAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["TxLossOfSignalFunctionalityAlarm"], prometheus.GaugeValue, opticLaneTxLossOfSignalFunctionalityAlarm, laneLabels...)

			opticLaneRxLossOfSignalAlarm := 0.0
			if lane.RxLossOfSignalAlarm.Text == "off" {
				opticLaneRxLossOfSignalAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["RxLossOfSignalAlarm"], prometheus.GaugeValue, opticLaneRxLossOfSignalAlarm, laneLabels...)

			opticLaneTxLaserDisabledAlarm := 0.0
			if lane.TxLaserDisabledAlarm.Text == "off" {
				opticLaneTxLaserDisabledAlarm = 1.0
			}
			ch <- prometheus.MustNewConstMetric(opticsDesc["TxLaserDisabledAlarm"], prometheus.GaugeValue, opticLaneTxLaserDisabledAlarm, laneLabels...)
		}
	}
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
