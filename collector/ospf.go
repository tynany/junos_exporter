package collector

import (
	"encoding/xml"
	"fmt"
	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
  ospfSubsystem   = "ospf"
  totalOSPFErrors = 0.0

  ospfPeerLabels = []string{"neighbor_address", "neighbor_id", "local_interface"}
  ospfDesc = map[string]*prometheus.Desc{
        "NeighborStatus": colPromDesc(ospfSubsystem, "neighbot_status", "OSPF Neighbor Status", ospfPeerLabels),   
	}
)

// EnvCollector collects environment metrics, implemented as per the Collector interface.
type OSPFCollector struct{}

// NewEnvCollector returns a new EnvCollector.
func NewOSPFCollector() *OSPFCollector {
	return &OSPFCollector{}
}

// Name of the collector.
func (*OSPFCollector) Name() string {
	return ospfSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *OSPFCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalOSPFErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalOSPFErrors
	}
	defer s.Close()

  // show ospf neighbor | display xml
	reply, err := s.Exec(netconf.RawMethod(`<get-ospf-neighbor-information/>`))
	if err != nil {
		totalOSPFErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalOSPFErrors
	}

    if err := processOSPFNetconfReply(reply, ch); err != nil {
        totalOSPFErrors++
        errors = append(errors, err)
    }
	return errors, totalOSPFErrors
}

func processOSPFNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
    ospfNbrStatus := 0.0
    var netconfReply ospfNeighborRPCReply
    if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
        return fmt.Errorf("could not unmarshal netconf ospf neighbor reply: %s", err)
    }
    for _, neighbor := range netconfReply.OSPFNbrInformation.OSPFNeighbor {
        ospfNbrStatus = 0.0
        ospfPeerLabels := []string{
            neighbor.NeighborAddress,
            neighbor.NeighborId,
            neighbor.InterfaceName,
        }
        if neighbor.OSPFNeighborState == "Full" {
            ospfNbrStatus = 1.0
        }
        ch <- prometheus.MustNewConstMetric(ospfDesc["NeighborStatus"], prometheus.GaugeValue, ospfNbrStatus, ospfPeerLabels...) 
    }
    return nil
}

type ospfNeighborRPCReply struct {
    XMLName             xml.Name            `xml:"rpc-reply"`
    Xmlns               string              `xml:"xmlns,attr"`
    OSPFNbrInformation  ospfNbrInformation  `xml:"ospf-neighbor-information"`
}

type ospfNbrInformation struct {
    OSPFNeighbor    []ospfNeighbor  `xml:"ospf-neighbor"`
}

type ospfNeighbor struct {
    NeighborAddress     string  `xml:"neighbor-address"`
    InterfaceName       string  `xml:"interface-name"`
    OSPFNeighborState   string  `xml:"ospf-neighbor-state"`
    NeighborId          string  `xml:"neighbor-id"`
    NeighborPriority    string  `xml:"neighbor-priority"`
    ActivityTimer       string  `xml:"activity-timer"`
}

