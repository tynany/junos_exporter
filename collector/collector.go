package collector

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"golang.org/x/crypto/ssh"
)

// The namespace used by all metrics.
const namespace = "junos"

var (
	junosTotalScrapeCount = 0.0
	junosTotalErrorCount  = 0
	junosLabels           = []string{"collector"}
	junosDesc             = map[string]*prometheus.Desc{
		"ScrapesTotal":   promDesc("scrapes_total", "Total number of times Junos has been scraped.", nil),
		"ScrapeErrTotal": promDesc("scrape_errors_total", "Total number of errors from a collector.", junosLabels),
		"ScrapeDuration": promDesc("scrape_duration_seconds", "Time it took for a collector's scrape to complete.", junosLabels),
		"CollectorUp":    promDesc("collector_up", "Whether the collector's last scrape was successful (1 = successful, 0 = unsuccessful).", junosLabels),
		"Up":             promDesc("up", "Whether the Junos collector is currently up.", nil),
	}

	sshClientConfig *ssh.ClientConfig
	sshTarget       string
	ifaceDescrKeys  []string
)

// CollectErrors is used to collect collector errors.
type CollectErrors interface {
	// Returns any errors that were encounted during Collect.
	CollectErrors() []error

	// Returns the total number of errors encounter during app run duration.
	CollectTotalErrors() float64
}

// SSHConfig contains the credentials required to create a *ssh.ClientConfig that is used to connect to a device.
type SSHConfig struct {
	Username string
	Timeout  time.Duration
	Password string
	SSHKey   []byte
}

// Exporters contains a slice of Collectors.
type Exporters struct {
	Collectors []*Collector
}

// Collector contains everything needed to collect from a collector.
type Collector struct {
	Name          string
	PromCollector prometheus.Collector
	Errors        CollectErrors
}

// NewExporter returns an Exporters type containing a slice of Collectors.
func NewExporter(collectors []*Collector) *Exporters {
	return &Exporters{
		Collectors: collectors,
	}
}

// SetConnectionDetails sets the sshClientConfig and sshTarget variables required to make a connection.
func (*Exporters) SetConnectionDetails(config SSHConfig, target string) error {
	sshTarget = target
	sshClientConfig = &ssh.ClientConfig{
		User: config.Username,
	}
	if len(config.SSHKey) > 0 {
		parsedKey, err := ssh.ParsePrivateKey(config.SSHKey)
		if err != nil {
			return err
		}
		sshClientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(parsedKey)}
	} else {
		sshClientConfig.Auth = []ssh.AuthMethod{ssh.Password(config.Password)}
	}
	sshClientConfig.Timeout = config.Timeout
	sshClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	return nil
}

// SetIfaceDescrKeys sets the optional keys in an interface description to include in metrics
func (*Exporters) SetIfaceDescrKeys(keys []string) {
	ifaceDescrKeys = keys
}

// Describe implemented as per the prometheus.Collector interface.
func (e *Exporters) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range bgpDesc {
		ch <- desc
	}
	for _, collector := range e.Collectors {
		collector.PromCollector.Describe(ch)
	}
}

// Collect implemented as per the prometheus.Collector interface.
func (e *Exporters) Collect(ch chan<- prometheus.Metric) {
	junosTotalScrapeCount++
	ch <- prometheus.MustNewConstMetric(junosDesc["ScrapesTotal"], prometheus.CounterValue, junosTotalScrapeCount)

	wg := &sync.WaitGroup{}
	for _, collector := range e.Collectors {
		wg.Add(1)
		go runCollector(ch, collector, wg)
	}
	wg.Wait()
}

func runCollector(ch chan<- prometheus.Metric, collector *Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	startTime := time.Now()
	collector.PromCollector.Collect(ch)
	ch <- prometheus.MustNewConstMetric(junosDesc["ScrapeErrTotal"], prometheus.GaugeValue, collector.Errors.CollectTotalErrors(), collector.Name)
	errors := collector.Errors.CollectErrors()
	if len(errors) > 0 {
		ch <- prometheus.MustNewConstMetric(junosDesc["CollectorUp"], prometheus.GaugeValue, 0, collector.Name)
		for _, err := range errors {
			log.Errorf("collector %q scrape failed: %s", collector.Name, err)
		}
	} else {
		ch <- prometheus.MustNewConstMetric(junosDesc["CollectorUp"], prometheus.GaugeValue, 1, collector.Name)
	}
	ch <- prometheus.MustNewConstMetric(junosDesc["ScrapeDuration"], prometheus.GaugeValue, float64(time.Since(startTime).Seconds()), collector.Name)
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
