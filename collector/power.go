package collector

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	powerSubsystem = "power"

	powerLabel           = []string{"module"}
	powerModuleZoneLabel = []string{"module", "zone"}
	powerZoneLabels      = []string{"zone"}
	powerDesc            = map[string]*prometheus.Desc{
		"State":                 colPromDesc(powerSubsystem, "module_state", "Module Power State (1 = Online, 0 = Offline).", powerLabel),
		"CapacityActual":        colPromDesc(powerSubsystem, "module_capacity_actual_watts", "Module Actual Capacity in Watts", powerLabel),
		"CapacityMax":           colPromDesc(powerSubsystem, "module_capacity_maximum_watts", "Module Maximum Capacity in Watts", powerLabel),
		"ACInputState":          colPromDesc(powerSubsystem, "module_ac_input_state", "Module AC Input State (1 = OK, 0 = Not OK).", powerLabel),
		"ACExpectedFeeds":       colPromDesc(powerSubsystem, "module_ac_input_expected_feeds", "Module AC Input Expected Feeds.", powerLabel),
		"ACConnectedFeeds":      colPromDesc(powerSubsystem, "module_ac_input_connected_feeds", "Module AC Input Connected Feeds.", powerLabel),
		"DCOutputPower":         colPromDesc(powerSubsystem, "module_dc_output_watts", "Module DC Output Watts (Power).", powerModuleZoneLabel),
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

	totalPowerErrors = 0.0
)

// PowerCollector collects power metrics, implemented as per the Collector interface.
type PowerCollector struct {
	logger log.Logger
}

// NewPowerCollector returns a new PowerCollector .
func NewPowerCollector(logger log.Logger) *PowerCollector {
	return &PowerCollector{logger: logger}
}

// Name of the collector.
func (*PowerCollector) Name() string {
	return powerSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *PowerCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalPowerErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalPowerErrors
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-power-usage-information-detail></get-power-usage-information-detail>`))
	if err != nil {
		totalPowerErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalPowerErrors
	}

	if err := processPowerNetconfReply(reply, ch, c.logger); err != nil {
		totalPowerErrors++
		errors = append(errors, err)
	}
	return errors, totalPowerErrors
}

func processPowerNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric, logger log.Logger) error {
	var netconfReply powerRPCReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, powerData := range netconfReply.PowerUsageInformation.PowerUsageItem {
		labels := []string{strings.TrimSpace(powerData.Name.Text)}

		powerState := 0.0
		if powerData.State.Text == "Online" {
			powerState = 1.0
		}
		ch <- prometheus.MustNewConstMetric(powerDesc["State"], prometheus.GaugeValue, powerState, labels...)
		newGauge(logger, ch, powerDesc["CapacityActual"], powerData.PemCapacityDetail.CapacityActual.Text, labels...)
		newGauge(logger, ch, powerDesc["CapacityMax"], powerData.PemCapacityDetail.CapacityMax.Text, labels...)

		powerACInputState := 0.0
		if powerData.AcInputDetail.AcInput.Text == "OK" {
			powerACInputState = 1.0
		}
		ch <- prometheus.MustNewConstMetric(powerDesc["ACInputState"], prometheus.GaugeValue, powerACInputState, labels...)
		newGauge(logger, ch, powerDesc["ACExpectedFeeds"], powerData.AcInputDetail.AcExpectFeed.Text, labels...)
		newGauge(logger, ch, powerDesc["ACConnectedFeeds"], powerData.AcInputDetail.AcActualFeed.Text, labels...)

		labelsDCOutput := []string{strings.TrimSpace(powerData.Name.Text), strings.TrimSpace(powerData.DcOutputDetail.Zone.Text)}
		newGauge(logger, ch, powerDesc["DCOutputPower"], powerData.DcOutputDetail.DcPower.Text, labelsDCOutput...)
		newGauge(logger, ch, powerDesc["DCOutputCurrent"], powerData.DcOutputDetail.DcCurrent.Text, labelsDCOutput...)
		newGauge(logger, ch, powerDesc["DCOutputVoltage"], powerData.DcOutputDetail.DcVoltage.Text, labelsDCOutput...)
		newGauge(logger, ch, powerDesc["DCOutputLoad"], powerData.DcOutputDetail.DcLoad.Text, labelsDCOutput...)
	}
	for _, powerSystem := range netconfReply.PowerUsageInformation.PowerUsageSystem {
		for _, zone := range powerSystem.PowerUsageZoneInformation {
			labels := []string{strings.TrimSpace(zone.Zone.Text)}
			newGauge(logger, ch, powerDesc["CapacityZoneActual"], zone.CapacityActual.Text, labels...)
			newGauge(logger, ch, powerDesc["CapacityZoneMax"], zone.CapacityMax.Text, labels...)
			newGauge(logger, ch, powerDesc["CapacityZoneAllocated"], zone.CapacityAllocated.Text, labels...)
			newGauge(logger, ch, powerDesc["CapacityZoneRemaining"], zone.CapacityRemaining.Text, labels...)
			newGauge(logger, ch, powerDesc["CapacityZoneUsage"], zone.CapacityActualUsage.Text, labels...)
		}
		newGauge(logger, ch, powerDesc["CapacitySysActual"], powerSystem.CapacitySysActual.Text)
		newGauge(logger, ch, powerDesc["CapacitySysMax"], powerSystem.CapacitySysMax.Text)
		newGauge(logger, ch, powerDesc["CapacitySysRemaining"], powerSystem.CapacitySysRemaining.Text)
	}
	for _, fruItem := range netconfReply.PowerUsageInformation.PowerUsageFruItem {
		newGauge(logger, ch, powerDesc["DCUsage"], fruItem.DcPower.Text, fruItem.Name.Text)
	}
	return nil
}

type powerRPCReply struct {
	PowerUsageInformation powerInformation `xml:"power-usage-information"`
}
type powerInformation struct {
	PowerUsageItem    []powerItem    `xml:"power-usage-item"`
	PowerUsageSystem  []powerSystem  `xml:"power-usage-system"`
	PowerUsageFruItem []powerFRUItem `xml:"power-usage-fru-item"`
}

type powerFRUItem struct {
	Name    powerText `xml:"name"`
	DcPower powerText `xml:"dc-power"`
}

type powerSystem struct {
	PowerUsageZoneInformation []powerZone `xml:"power-usage-zone-information"`
	CapacitySysActual         powerText   `xml:"capacity-sys-actual"`
	CapacitySysMax            powerText   `xml:"capacity-sys-max"`
	CapacitySysRemaining      powerText   `xml:"capacity-sys-remaining"`
}

type powerZone struct {
	Zone                powerText `xml:"zone"`
	CapacityActual      powerText `xml:"capacity-actual"`
	CapacityMax         powerText `xml:"capacity-max"`
	CapacityAllocated   powerText `xml:"capacity-allocated"`
	CapacityRemaining   powerText `xml:"capacity-remaining"`
	CapacityActualUsage powerText `xml:"capacity-actual-usage"`
}

type powerItem struct {
	Name              powerText        `xml:"name"`
	State             powerText        `xml:"state"`
	PemCapacityDetail powerPEMCapacity `xml:"pem-capacity-detail"`
	AcInputDetail     powerACInput     `xml:"ac-input-detail"`
	DcOutputDetail    powerDCOutput    `xml:"dc-output-detail"`
}

type powerDCOutput struct {
	DcPower   powerText `xml:"dc-power"`
	Zone      powerText `xml:"zone"`
	DcCurrent powerText `xml:"dc-current"`
	DcVoltage powerText `xml:"dc-voltage"`
	DcLoad    powerText `xml:"dc-load"`
}

type powerACInput struct {
	AcInput      powerText `xml:"ac-input"`
	AcExpectFeed powerText `xml:"ac-expect-feed"`
	AcActualFeed powerText `xml:"ac-actual-feed"`
}

type powerPEMCapacity struct {
	CapacityActual powerText `xml:"capacity-actual"`
	CapacityMax    powerText `xml:"capacity-max"`
}
type powerText struct {
	Text string `xml:",chardata"`
}

// Leaving the below as it may be implemented in the future
// type powerTemp struct {
// 	Temp string `xml:"celsius,attr"`
// }
