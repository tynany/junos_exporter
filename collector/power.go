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

	powerErrors      = []error{}
	totalPowerErrors = 0.0
)

// PowerCollector collects power metrics, implemented as per the Collector power.
type PowerCollector struct{}

// NewPowerCollector returns a PowerCollector type.
func NewPowerCollector() *PowerCollector {
	return &PowerCollector{}
}

// Name of the collector. Used to parse the configuration file.
func (*PowerCollector) Name() string {
	return powerSubsystem
}

// CollectErrors returns what errors have been gathered.
func (*PowerCollector) CollectErrors() []error {
	errors := powerErrors
	powerErrors = []error{}
	return errors
}

// CollectTotalErrors collects total errors.
func (*PowerCollector) CollectTotalErrors() float64 {
	return totalPowerErrors
}

// Describe all metrics implemented as per the prometheus.Collector interface.
func (*PowerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range powerDesc {
		ch <- desc
	}
}

// Collect metrics as per the prometheus.Collector interface.
func (c *PowerCollector) Collect(ch chan<- prometheus.Metric) {
	s, err := netconf.DialSSH(sshTarget, sshClientConfig)
	if err != nil {
		totalPowerErrors++
		powerErrors = append(powerErrors, fmt.Errorf("could not connect to %q: %s", sshTarget, err))
		return
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-power-usage-information-detail></get-power-usage-information-detail>`))
	if err != nil {
		totalPowerErrors++
		powerErrors = append(powerErrors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return
	}

	if err := processPowerNetconfReply(reply, ch); err != nil {
		totalPowerErrors++
		powerErrors = append(powerErrors, err)
	}
}

func processPowerNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
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
		newGauge(ch, powerDesc["DCOutputPower"], powerData.DcOutputDetail.DcPower.Text, labelsDCOutput...)
		newGauge(ch, powerDesc["DCOutputCurrent"], powerData.DcOutputDetail.DcCurrent.Text, labelsDCOutput...)
		newGauge(ch, powerDesc["DCOutputVoltage"], powerData.DcOutputDetail.DcVoltage.Text, labelsDCOutput...)
		newGauge(ch, powerDesc["DCOutputLoad"], powerData.DcOutputDetail.DcLoad.Text, labelsDCOutput...)
	}
	for _, powerSystem := range netconfReply.PowerUsageInformation.PowerUsageSystem {
		for _, zone := range powerSystem.PowerUsageZoneInformation {
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
	for _, fruItem := range netconfReply.PowerUsageInformation.PowerUsageFruItem {
		newGauge(ch, powerDesc["DCUsage"], fruItem.DcPower.Text, fruItem.Name.Text)
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

type powerTemp struct {
	Temp string `xml:"celsius,attr"`
}
