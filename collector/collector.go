package collector

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/common/log"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/ssh"
)

// The namespace used by all metrics.
const namespace = "junos"

var (
	junosTotalScrapeCount = 0.0
	junosLabels           = []string{"collector"}
	junosDesc             = map[string]*prometheus.Desc{
		"ScrapesTotal":   promDesc("scrapes_total", "Total number of times Junos has been scraped.", nil),
		"ScrapeErrTotal": promDesc("scrape_errors_total", "Total number of errors from a collector.", junosLabels),
		"ScrapeDuration": promDesc("scrape_duration_seconds", "Time it took for a collector's scrape to complete.", junosLabels),
		"CollectorUp":    promDesc("collector_up", "Whether the collector's last scrape was successful (1 = successful, 0 = unsuccessful).", junosLabels),
		"Up":             promDesc("up", "Whether the Junos collector is currently up.", nil),
	}
)

// Collector is the interface a collector has to implement.
type Collector interface {
	// Returns the name of the collector.
	Name() string
	// Gets metrics and sends to the Prometheus.Metric channel.
	Get(ch chan<- prometheus.Metric, config Config) ([]error, float64)
}

// Config required by the collectors.
type Config struct {
	SSHClientConfig *ssh.ClientConfig
	SSHTarget       string
	IfaceDescrKeys  []string
	IfaceMetricKeys []string
	BGPTypeKeys     []string
}

// Exporter collects all exporter metrics, implemented as per the prometheus.Collector interface.
type Exporter struct {
	Collectors []Collector
	config     Config
}

// NewExporter returns a new Exporter.
func NewExporter(collectors []Collector, config Config) (*Exporter, error) {
	return &Exporter{
		Collectors: collectors,
		config:     config,
	}, nil
}

// Collect implemented as per the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	junosTotalScrapeCount++
	ch <- prometheus.MustNewConstMetric(junosDesc["ScrapesTotal"], prometheus.CounterValue, junosTotalScrapeCount)

	wg := &sync.WaitGroup{}
	for _, collector := range e.Collectors {
		wg.Add(1)
		go e.runCollector(ch, collector, wg)
	}
	wg.Wait()
}

func (e *Exporter) runCollector(ch chan<- prometheus.Metric, collector Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	collectorName := collector.Name()

	startTime := time.Now()
	errors, totalErrors := collector.Get(ch, e.config)

	ch <- prometheus.MustNewConstMetric(junosDesc["ScrapeDuration"], prometheus.GaugeValue, float64(time.Since(startTime).Seconds()), collectorName)
	ch <- prometheus.MustNewConstMetric(junosDesc["ScrapeErrTotal"], prometheus.GaugeValue, totalErrors, collectorName)

	if len(errors) > 0 {
		ch <- prometheus.MustNewConstMetric(junosDesc["CollectorUp"], prometheus.GaugeValue, 0, collector.Name())
		for _, err := range errors {
			log.Errorf("collector %q scrape failed: %s", collectorName, err)
		}
	} else {
		ch <- prometheus.MustNewConstMetric(junosDesc["CollectorUp"], prometheus.GaugeValue, 1, collectorName)
	}
}

// Describe implemented as per the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range junosDesc {
		ch <- desc
	}
}

func promDesc(metricName string, metricDescription string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, metricDescription, labels, nil)
}

func colPromDesc(subsystem string, metricName string, metricDescription string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, metricName), metricDescription, labels, nil)
}

func newGauge(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric string, labels ...string) {
	if metric != "" {
		i, err := strconv.ParseFloat(strings.TrimSpace(metric), 64)
		if err != nil {
			log.Errorf("could not convert metric to float64: %s", err)
		}
		ch <- prometheus.MustNewConstMetric(descName, prometheus.GaugeValue, i, labels...)
	}
}

func newCounter(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric string, labels ...string) {
	if metric != "" {
		i, err := strconv.ParseFloat(strings.TrimSpace(metric), 64)
		if err != nil {
			log.Errorf("could not convert metric to float64: %s", err)
		}
		ch <- prometheus.MustNewConstMetric(descName, prometheus.CounterValue, i, labels...)
	}
}

func newGaugeMB(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric string, labels ...string) {
	if metric != "" {
		re := regexp.MustCompile("[0-9]+")
		i, err := strconv.ParseFloat(strings.TrimSpace(re.FindString(metric)), 64)
		if err != nil {
			log.Errorf("could not convert metric to float64: %s", err)
		}

		ch <- prometheus.MustNewConstMetric(descName, prometheus.GaugeValue, i*1000000, labels...)
	}
}
