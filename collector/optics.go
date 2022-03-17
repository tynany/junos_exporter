package collector

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	powerSubsystem = "power"

	powerLabel           = []string{"module"}
	powerModuleZoneLabel = []string{"module", "zone"}
	powerZoneLabels      = []string{"zone"}
	powerDesc            = map[string]*prometheus.Desc{
		"State":                 colPromDesc(powerSubsystem, "module_state", "Module Optics State (1 = Online, 0 = Offline).", powerLabel),
		"CapacityActual":        colPromDesc(powerSubsystem, "module_capacity_actual_watts", "Module Actual Capacity in Watts", powerLabel),
		"CapacityMax":           colPromDesc(powerSubsystem, "module_capacity_maximum_watts", "Module Maximum Capacity in Watts", powerLabel),
		"ACInputState":          colPromDesc(powerSubsystem, "module_ac_input_state", "Module AC Input State (1 = OK, 0 = Not OK).", powerLabel),
		"ACExpectedFeeds":       colPromDesc(powerSubsystem, "module_ac_input_expected_feeds", "Module AC Input Expected Feeds.", powerLabel),
		"ACConnectedFeeds":      colPromDesc(powerSubsystem, "module_ac_input_connected_feeds", "Module AC Input Connected Feeds.", powerLabel),
		"DCOutputOptics":        colPromDesc(powerSubsystem, "module_dc_output_watts", "Module DC Output Watts (Optics).", powerModuleZoneLabel),
		"DCOutputCurrent":       colPromDesc(powerSubsystem, "module_dc_output_amperes", "Module DC Output Amps (Current).", powerModuleZoneLabel),
		"DCOutputVoltage":       colPromDesc(powerSubsystem, "module_dc_output_volts", "Module DC Output Volts (Voltage).", powerModuleZoneLabel),
		"DCOutputLoad":          colPromDesc(powerSubsystem, "module_dc_output_load_ratio", "Module DC Output Load as a Percent.", powerModuleZoneLabel),
		"CapacityZoneActual":    colPromDesc(powerSubsystem, "system_zone_capacity_actual_watts", "System Zone Actual Capacity in Watts", powerZoneLabels),
		"CapacityZoneMax":       colPromDesc(powerSubsystem, "system_zone_capacity_maximum_watts", "System Zone Maximum Capacity in Watts", powerZoneLabels),
		"CapacityZoneAllocated": colPromDesc(powerSubsystem, "system_zone_allocated_watts", "System Zone Allocated Capacity in Watts", powerZoneLabels),
		"CapacityZoneRemaining": colPromDesc(powerSubsystem, "system_zone_remaining_watts", "System Zone Remaining Capacity in Watts", powerZoneLabels),
		"CapacityZoneUsage":     colPromDesc(powerSubsystem, "system_zone_usage_watts", "System Zone Usage in Watts", powerZoneLabels),
		"CapacitySysActual":     colPromDesc(powerSubsystem, "system_capacity_actual_watts", "System Actual Capacity in Watts", nil),
		"CapacitySysMax":        colPromDesc(powerSubsystem, "system_capacity_maximum_watts", "System Maximum Capacity in Watts", nil),
		"CapacitySysRemaining":  colPromDesc(powerSubsystem, "system_remaining_watts", "System Remaining Capacity in Watts", nil),
		"DCUsage":               colPromDesc(powerSubsystem, "module_dc_usage_watts", "Module DC Usage in Watts.", powerLabel),
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
	for _, powerData := range netconfReply.OpticsUsageInformation.OpticsUsageItem {
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
	ModuleTemperature opticsTemp `xml:"module-temperature"`
	ModuleVoltage     opticsText `xml:"module-voltage"`
	/*
		`xml: module-temperature-high-alarm`
		`xml: module-temperature-low-alarm`
		`xml: module-temperature-high-warn`
		`xml: module-temperature-low-warn`
		`xml: module-voltage-high-alarm`
		`xml: module-voltage-low-alarm`
		`xml: module-voltage-high-warn`
		`xml: module-voltage-low-warn`
		`xml: module-temperature-high-alarm-threshold`
		`xml: module-temperature-low-alarm-threshold`
		`xml: /module-temperature-low-alarm-threshold`
		`xml: module-temperature-high-warn-threshold`
		`xml: module-temperature-low-warn-threshold`
		`xml: /module-temperature-low-warn-threshold`
		`xml: module-voltage-high-alarm-threshold`
		`xml: module-voltage-low-alarm-threshold`
		`xml: module-voltage-high-warn-threshold`
		`xml: module-voltage-low-warn-threshold`
		`xml: laser-bias-current-high-alarm-threshold`
		`xml: laser-bias-current-low-alarm-threshold`
		`xml: laser-bias-current-high-warn-threshold`
		`xml: laser-bias-current-low-warn-threshold`
		`xml: laser-tx-power-high-alarm-threshold`
		`xml: laser-tx-power-high-alarm-threshold-dbm`
		`xml: laser-tx-power-low-alarm-threshold`
		`xml: laser-tx-power-low-alarm-threshold-dbm`
		`xml: laser-tx-power-high-warn-threshold`
		`xml: laser-tx-power-high-warn-threshold-dbm`
		`xml: laser-tx-power-low-warn-threshold`
		`xml: laser-tx-power-low-warn-threshold-dbm`
		`xml: laser-rx-power-high-alarm-threshold`
		`xml: laser-rx-power-high-alarm-threshold-dbm`
		`xml: laser-rx-power-low-alarm-threshold`
		`xml: laser-rx-power-low-alarm-threshold-dbm`
		`xml: laser-rx-power-high-warn-threshold`
		`xml: laser-rx-power-high-warn-threshold-dbm`
		`xml: laser-rx-power-low-warn-threshold`
		`xml: laser-rx-power-low-warn-threshold-dbm`
	*/

	OpticsDiagLanes []OpticsDiagLanes `xml:"optics-diagnostics-lane-values"`
}

type OpticsDiagLanes struct {
	LaneIndex opticsText `xml:"lane-index"`
	LaneRXDbm opticsText `xml:"laser-rx-optical-power-dbm"`
	LaneTXDbm opticsText `xml:"laser-output-power-dbm"`
}

type opticsText struct {
	Text string `xml:",chardata"`
}

type opticsTemp struct {
	Temp string `xml:"celsius,attr"`
}
