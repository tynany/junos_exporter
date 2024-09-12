package collector

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	fpcSubsystem = "fpc"

	fpcLabels    = []string{"slot"}
	fpcCPULabels = append(fpcLabels, "timespan")

	fpcDesc = map[string]*prometheus.Desc{
		"State":            colPromDesc(fpcSubsystem, "state", "State (0 = Offline, 1 = Online, 2 = Empty, 4 = Other).", fpcLabels),
		"Temp":             colPromDesc(fpcSubsystem, "temperature_celsius", "Temperature in Celsius", fpcLabels),
		"CPUTotal":         colPromDesc(fpcSubsystem, "cpu_total", "Total CPU utilization.", fpcLabels),
		"CPUInterrupt":     colPromDesc(fpcSubsystem, "cpu_interrupt", "CPU Interrupt utilization.", fpcLabels),
		"CPUAvg":           colPromDesc(fpcSubsystem, "cpu_avg", "Average CPU utilization across timespan.", fpcCPULabels),
		"MemoryDramSize":   colPromDesc(fpcSubsystem, "memory_dram_size", "Memory DRAM Size.", fpcLabels),
		"MemoryHeapUtil":   colPromDesc(fpcSubsystem, "memory_heap_utilization", "Memory heap utilization.", fpcLabels),
		"MemoryBufferUtil": colPromDesc(fpcSubsystem, "memory_buffer_utilization", "Memory buffer utilization.", fpcLabels),
	}

	totalFPCErrors = 0.0
)

// FPCCollector collects environment metrics, implemented as per the Collector interface.
type FPCCollector struct{}

// NewFPCCollector returns a new FPCCollector.
func NewFPCCollector() *FPCCollector {
	return &FPCCollector{}
}

// Name of the collector.
func (*FPCCollector) Name() string {
	return fpcSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *FPCCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalFPCErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalFPCErrors
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-fpc-information/>`))
	if err != nil {
		totalFPCErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalFPCErrors
	}

	if err := processFPCNetconfReply(reply, ch); err != nil {
		totalFPCErrors++
		errors = append(errors, err)
	}

	return errors, totalFPCErrors
}

func processFPCNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
	var netconfReply fpcRPCReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, data := range netconfReply.FPCInformation.FPCItem {
		labels := []string{strings.TrimSpace(data.Slot)}
		state := 3.0
		if strings.ToLower(data.State) == "offline" {
			state = 0.0
		} else if strings.ToLower(data.State) == "online" {
			state = 1.0
			ch <- prometheus.MustNewConstMetric(fpcDesc["Temp"], prometheus.GaugeValue, data.Temperature, labels...)
			ch <- prometheus.MustNewConstMetric(fpcDesc["CPUTotal"], prometheus.GaugeValue, data.CPUTotal, labels...)
			ch <- prometheus.MustNewConstMetric(fpcDesc["CPUInterrupt"], prometheus.GaugeValue, data.CPUInterrupt, labels...)

			label1m := append(labels, "1m")
			ch <- prometheus.MustNewConstMetric(fpcDesc["CPUAvg"], prometheus.GaugeValue, data.CPU1MAvg, label1m...)

			label5m := append(labels, "5m")
			ch <- prometheus.MustNewConstMetric(fpcDesc["CPUAvg"], prometheus.GaugeValue, data.CPU5MinAvg, label5m...)

			label15m := append(labels, "15m")
			ch <- prometheus.MustNewConstMetric(fpcDesc["CPUAvg"], prometheus.GaugeValue, data.CPU15MinAvg, label15m...)

			ch <- prometheus.MustNewConstMetric(fpcDesc["MemoryDramSize"], prometheus.GaugeValue, data.MemoryDRAMSize, labels...)
			ch <- prometheus.MustNewConstMetric(fpcDesc["MemoryHeapUtil"], prometheus.GaugeValue, data.MemoryHeapUtilization, labels...)
			ch <- prometheus.MustNewConstMetric(fpcDesc["MemoryBufferUtil"], prometheus.GaugeValue, data.MemoryBufferUtilization, labels...)
		} else if strings.ToLower(data.State) == "empty" {
			state = 3.0
		}
		ch <- prometheus.MustNewConstMetric(fpcDesc["State"], prometheus.GaugeValue, state, labels...)
	}
	return nil
}

type fpcRPCReply struct {
	FPCInformation fpcInformation `xml:"fpc-information"`
}

type fpcInformation struct {
	FPCItem []fpcItem `xml:"fpc"`
}

type fpcItem struct {
	Slot                    string  `xml:"slot"`
	State                   string  `xml:"state"`
	Temperature             float64 `xml:"temperature"`
	CPUTotal                float64 `xml:"cpu-total"`
	CPUInterrupt            float64 `xml:"cpu-interrupt"`
	CPU1MAvg                float64 `xml:"cpu-1min-avg"`
	CPU5MinAvg              float64 `xml:"cpu-5min-avg"`
	CPU15MinAvg             float64 `xml:"cpu-15min-avg"`
	MemoryDRAMSize          float64 `xml:"memory-dram-size"`
	MemoryHeapUtilization   float64 `xml:"memory-heap-utilization"`
	MemoryBufferUtilization float64 `xml:"memory-buffer-utilization"`
}
