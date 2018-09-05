package collector

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	reSubsystem = "route_engine"

	reLabels    = []string{"slot"}
	reCPULabels = append(reLabels, "timespan")
	reDesc      = map[string]*prometheus.Desc{
		"state":         colPromDesc(reSubsystem, "state", "RE state (1 = OK, 0 = Not OK).", reLabels),
		"temp":          colPromDesc(reSubsystem, "temperature_celsius", "Route engine temperature in degrees celsius.", reLabels),
		"cpuTemp":       colPromDesc(reSubsystem, "cpu_temperature_celsius", "Route engine CPU temperature in degrees celsius.", reLabels),
		"memTotal":      colPromDesc(reSubsystem, "memory_total_bytes", "Total route engine memory in bytes.", reLabels),
		"memUsed":       colPromDesc(reSubsystem, "memory_used_bytes", "Used route engine memory in bytes.", reLabels),
		"memBuf":        colPromDesc(reSubsystem, "memory_buffer_utilization_percent", "Memory buffer utilization as a percent.", reLabels),
		"memDRAM":       colPromDesc(reSubsystem, "memory_dram_size_bytes", "Memory DRAM size in bytes.", reLabels),
		"memInstalled":  colPromDesc(reSubsystem, "memory_installed_size_bytes", "Memory installed size in bytes.", reLabels),
		"cpuUser":       colPromDesc(reSubsystem, "cpu_user_percent", "User CPU utilization as a percent.", reCPULabels),
		"cpuBackground": colPromDesc(reSubsystem, "cpu_background_percent", "Background CPU utilization as a percent.", reCPULabels),
		"cpuSystem":     colPromDesc(reSubsystem, "cpu_system_percent", "System CPU utilization as a percent.", reCPULabels),
		"cpuInterrupt":  colPromDesc(reSubsystem, "cpu_interrupt_percent", "Interrupt CPU utilization as a percent.", reCPULabels),
		"cpuIdle":       colPromDesc(reSubsystem, "cpu_idle_percent", "Idle CPU utilization as a percent.", reCPULabels),
		"loadAvg":       colPromDesc(reSubsystem, "load_average", "LoadAverage.", reCPULabels),
		"uptime":        colPromDesc(reSubsystem, "uptime_seconds", "Uptime in seconds.", reLabels),
		"masterState":   colPromDesc(reSubsystem, "mastership_state", "Mastership state (1 = Master, 0 = Backup).", reLabels),
		"masterPrio":    colPromDesc(reSubsystem, "mastership_priority", "Mastership priority (1 = Master, 0 = Backup).", reLabels),
	}

	reErrors      = []error{}
	totalREErrors = 0.0
)

// RECollector collects route enginer metrics.
type RECollector struct{}

// NewRECollector returns a RECollector type.
func NewRECollector() *RECollector {
	return &RECollector{}
}

// Name of the collector. Used to parse the configuration file.
func (*RECollector) Name() string {
	return reSubsystem
}

// CollectErrors returns what errors have been gathered.
func (*RECollector) CollectErrors() []error {
	errors := reErrors
	reErrors = []error{}
	return errors
}

// CollectTotalErrors collects total errors.
func (*RECollector) CollectTotalErrors() float64 {
	return totalREErrors
}

// Describe all metrics implemented as per the prometheus.Collector interface.
func (*RECollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range reDesc {
		ch <- desc
	}
}

// Collect metrics as per the prometheus.Collector interface.
func (c *RECollector) Collect(ch chan<- prometheus.Metric) {
	s, err := netconf.DialSSH(sshTarget, sshClientConfig)
	if err != nil {
		totalREErrors++
		reErrors = append(reErrors, fmt.Errorf("could not connect to %q: %s", sshTarget, err))
		return
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-route-engine-information/>`))
	if err != nil {
		totalREErrors++
		reErrors = append(reErrors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return
	}

	if err := processRENetconfReply(reply, ch); err != nil {
		totalREErrors++
		reErrors = append(reErrors, err)
	}
}

func processRENetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
	var netconfReply reRPCReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}

	for _, reData := range netconfReply.REInformation.REEntry {
		var labels []string
		if reData.Slot.Text != "" {
			labels = append(labels, reData.Slot.Text)
		} else {
			labels = append(labels, "singleRE")
		}

		state := 0.0
		if strings.ToLower(reData.Status.Text) == "ok" {
			state = 1.0
		}
		ch <- prometheus.MustNewConstMetric(reDesc["state"], prometheus.GaugeValue, state, labels...)

		newGauge(ch, reDesc["temp"], reData.Temperature.Temp, labels...)
		newGauge(ch, reDesc["cpuTemp"], reData.CPUTemperature.Temp, labels...)
		newGauge(ch, reDesc["uptime"], reData.UpTime.Seconds, labels...)

		newGuageMB(ch, reDesc["memTotal"], reData.MemorySystemTotal.Text, labels...)
		newGuageMB(ch, reDesc["memUsed"], reData.MemorySystemTotalUsed.Text, labels...)
		newGauge(ch, reDesc["memBuf"], reData.MemoryBufferUtilization.Text, labels...)
		newGuageMB(ch, reDesc["memDRAM"], reData.MemoryDRAMSize.Text, labels...)
		newGuageMB(ch, reDesc["memInstalled"], reData.MemoryInstalledSize.Text, labels...)

		label5s := append(labels, "5s")
		newGauge(ch, reDesc["cpuUser"], reData.CPUUser.Text, label5s...)
		newGauge(ch, reDesc["cpuBackground"], reData.CPUBackground.Text, label5s...)
		newGauge(ch, reDesc["cpuSystem"], reData.CPUSystem.Text, label5s...)
		newGauge(ch, reDesc["cpuInterrupt"], reData.CPUInterrupt.Text, label5s...)
		newGauge(ch, reDesc["cpuIdle"], reData.CPUIdle.Text, label5s...)

		label1m := append(labels, "1m")
		newGauge(ch, reDesc["cpuUser"], reData.CPUUser1.Text, label1m...)
		newGauge(ch, reDesc["cpuBackground"], reData.CPUBackground1.Text, label1m...)
		newGauge(ch, reDesc["cpuSystem"], reData.CPUSystem1.Text, label1m...)
		newGauge(ch, reDesc["cpuInterrupt"], reData.CPUInterrupt1.Text, label1m...)
		newGauge(ch, reDesc["cpuIdle"], reData.CPUIdle1.Text, label1m...)
		newGauge(ch, reDesc["loadAvg"], reData.LoadAverageOne.Text, label1m...)

		label5m := append(labels, "5m")
		newGauge(ch, reDesc["cpuUser"], reData.CPUUser2.Text, label5m...)
		newGauge(ch, reDesc["cpuBackground"], reData.CPUBackground2.Text, label5m...)
		newGauge(ch, reDesc["cpuSystem"], reData.CPUSystem2.Text, label5m...)
		newGauge(ch, reDesc["cpuInterrupt"], reData.CPUInterrupt2.Text, label5m...)
		newGauge(ch, reDesc["cpuIdle"], reData.CPUIdle2.Text, label5m...)
		newGauge(ch, reDesc["loadAvg"], reData.LoadAverageFive.Text, label5m...)

		label15m := append(labels, "15m")
		newGauge(ch, reDesc["cpuUser"], reData.CPUUser3.Text, label15m...)
		newGauge(ch, reDesc["cpuBackground"], reData.CPUBackground3.Text, label15m...)
		newGauge(ch, reDesc["cpuSystem"], reData.CPUSystem3.Text, label15m...)
		newGauge(ch, reDesc["cpuInterrupt"], reData.CPUInterrupt3.Text, label15m...)
		newGauge(ch, reDesc["cpuIdle"], reData.CPUIdle3.Text, label15m...)
		newGauge(ch, reDesc["loadAvg"], reData.LoadAverageFifteen.Text, label15m...)

		mState := 0.0
		if strings.ToLower(reData.MastershipState.Text) == "master" {
			mState = 1.0
		}
		ch <- prometheus.MustNewConstMetric(reDesc["masterState"], prometheus.GaugeValue, mState, labels...)

		mPri := 0.0
		if strings.Contains(strings.ToLower(reData.MastershipState.Text), "master") {
			mPri = 1.0
		}
		ch <- prometheus.MustNewConstMetric(reDesc["masterPrio"], prometheus.GaugeValue, mPri, labels...)

	}
	return nil
}

type reRPCReply struct {
	REInformation reInformation `xml:"route-engine-information"`
}

type reInformation struct {
	REEntry []reEntry `xml:"route-engine"`
}

type reEntry struct {
	Slot                    reText    `xml:"slot"`
	MastershipState         reText    `xml:"mastership-state"`
	MastershipPriority      reText    `xml:"mastership-priority"`
	Status                  reText    `xml:"status"`
	Temperature             reTemp    `xml:"temperature"`
	CPUTemperature          reTemp    `xml:"cpu-temperature"`
	MemoryDRAMSize          reText    `xml:"memory-dram-size"`
	MemoryInstalledSize     reText    `xml:"memory-installed-size"`
	MemoryBufferUtilization reText    `xml:"memory-buffer-utilization"`
	MemorySystemTotal       reText    `xml:"memory-system-total"`
	MemorySystemTotalUsed   reText    `xml:"memory-system-total-used"`
	CPUUser                 reText    `xml:"cpu-user"`
	CPUBackground           reText    `xml:"cpu-background"`
	CPUSystem               reText    `xml:"cpu-system"`
	CPUInterrupt            reText    `xml:"cpu-interrupt"`
	CPUIdle                 reText    `xml:"cpu-idle"`
	CPUUser1                reText    `xml:"cpu-user1"`
	CPUBackground1          reText    `xml:"cpu-background1"`
	CPUSystem1              reText    `xml:"cpu-system1"`
	CPUInterrupt1           reText    `xml:"cpu-interrupt1"`
	CPUIdle1                reText    `xml:"cpu-idle1"`
	CPUUser2                reText    `xml:"cpu-user2"`
	CPUBackground2          reText    `xml:"cpu-background2"`
	CPUSystem2              reText    `xml:"cpu-system2"`
	CPUInterrupt2           reText    `xml:"cpu-interrupt2"`
	CPUIdle2                reText    `xml:"cpu-idle2"`
	CPUUser3                reText    `xml:"cpu-user3"`
	CPUBackground3          reText    `xml:"cpu-background3"`
	CPUSystem3              reText    `xml:"cpu-system3"`
	CPUInterrupt3           reText    `xml:"cpu-interrupt3"`
	CPUIdle3                reText    `xml:"cpu-idle3"`
	UpTime                  reSeconds `xml:"up-time"`
	LoadAverageOne          reText    `xml:"load-average-one"`
	LoadAverageFive         reText    `xml:"load-average-five"`
	LoadAverageFifteen      reText    `xml:"load-average-fifteen"`
}

type reText struct {
	Text string `xml:",chardata"`
}

type reTemp struct {
	Temp string `xml:"celsius,attr"`
}

type reSeconds struct {
	Seconds string `xml:"seconds,attr"`
}
