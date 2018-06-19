package collector

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	bgpSubsystem = "bgp"

	bgpPeerLabels     = []string{"peer"}
	bgpRIBLabels      = []string{"address_family"}
	bgpPeerRIBLabels  = append(bgpPeerLabels, bgpRIBLabels...)
	bgpPeerTypeLabels = []string{"type"}
	bgpNoLabels       = []string{}
	bgpDesc           = map[string]*prometheus.Desc{
		"GroupCount":                       colPromDesc(bgpSubsystem, "groups", "Number of Configured Groups.", bgpNoLabels),
		"PeerCount":                        colPromDesc(bgpSubsystem, "peers", "Number of Configured Peers.", bgpNoLabels),
		"DownPeerCount":                    colPromDesc(bgpSubsystem, "down_peers", "Number of Peers that are Down.", bgpNoLabels),
		"PeerInputMessages":                colPromDesc(bgpSubsystem, "peer_input_messages", "Number of Input Messages for a Peer.", bgpPeerLabels),
		"PeerOutputMessages":               colPromDesc(bgpSubsystem, "peer_output_messages", "Number of Output Messages for a Peer.", bgpPeerLabels),
		"PeerRouteQueueCount":              colPromDesc(bgpSubsystem, "peer_route_queue", "Number of Route Queues for a Peer.", bgpPeerLabels),
		"PeerFlapCount":                    colPromDesc(bgpSubsystem, "peer_flaps", "Number of Time the Peer has Flapped.", bgpPeerLabels),
		"PeerElapsedTime":                  colPromDesc(bgpSubsystem, "peer_elapsed_time_seconds", "Length of Time the Peer has Been Up.", bgpPeerLabels),
		"PeerPeerState":                    colPromDesc(bgpSubsystem, "peer_up", "State of the Peer. (1 = Established, 0 = Down).", bgpPeerLabels),
		"PeerRIBActivePrefixCount":         colPromDesc(bgpSubsystem, "peer_rib_active_prefixes", "Number of Active Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBReceivedPrefixCount":       colPromDesc(bgpSubsystem, "peer_rib_received_prefixes", "Number of Received Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBAcceptedPrefixCount":       colPromDesc(bgpSubsystem, "peer_rib_accepted_prefixes", "Number of Accepted Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBSuppressedPrefixCount":     colPromDesc(bgpSubsystem, "peer_rib_suppressed_prefixes", "Number of Suppressed Prefixes for the Peer.", bgpPeerRIBLabels),
		"RIBTotalPrefixCount":              colPromDesc(bgpSubsystem, "rib_total_prefixes", "Total Number of Prefixes in the RIB.", bgpRIBLabels),
		"RIBReceivedPrefixCount":           colPromDesc(bgpSubsystem, "rib_received_prefixes", "Number of Received Prefixes in the RIB.", bgpRIBLabels),
		"RIBAcceptedPrefixCount":           colPromDesc(bgpSubsystem, "rib_accepted_prefixes", "Number of Accepted Prefixes in the RIB.", bgpRIBLabels),
		"RIBActivePrefixCount":             colPromDesc(bgpSubsystem, "rib_active_prefixes", "Number of Active Prefixes in the RIB.", bgpRIBLabels),
		"RIBSuppressedPrefixCount":         colPromDesc(bgpSubsystem, "rib_suppressed_prefixes", "Number of Suppressed Prefixes in the RIB.", bgpRIBLabels),
		"RIBHistoryPrefixCount":            colPromDesc(bgpSubsystem, "rib_history_prefixes", "History Prefix Count in the RIB.", bgpRIBLabels),
		"RIBDampedPrefixCount":             colPromDesc(bgpSubsystem, "rib_damped_prefixes", "Number of Dampened Prefixes in the RIB.", bgpRIBLabels),
		"RIBTotalExternalPrefixCount":      colPromDesc(bgpSubsystem, "rib_total_external_prefixes", "Total Number of External Prefixes in the RIB.", bgpRIBLabels),
		"RIBActiveExternalPrefixCount":     colPromDesc(bgpSubsystem, "rib_active_external_prefixes", "Number of Active External Prefixes in the RIB.", bgpRIBLabels),
		"RIBAcceptedExternalPrefixCount":   colPromDesc(bgpSubsystem, "rib_accepted_external_prefixes", "Number of Accepted External Prefixes in the RIB.", bgpRIBLabels),
		"RIBSuppressedExternalPrefixCount": colPromDesc(bgpSubsystem, "rib_suppressed_external_prefixes", "Number of Suppressed External Prefixes in the RIB.", bgpRIBLabels),
		"RIBTotalInternalPrefixCount":      colPromDesc(bgpSubsystem, "rib_total_internal_prefixes", "Total Number of Internal Prefixes in the RIB.", bgpRIBLabels),
		"RIBActiveInternalPrefixCount":     colPromDesc(bgpSubsystem, "rib_active_internal_prefixes", "Number of Active Internal Prefixes in the RIB.", bgpRIBLabels),
		"RIBAcceptedInternalPrefixCount":   colPromDesc(bgpSubsystem, "rib_accepted_internal_prefixes", "Number of Accepted Internal Prefixes in the RIB.", bgpRIBLabels),
		"RIBSuppressedInternalPrefixCount": colPromDesc(bgpSubsystem, "rib_suppressed_internal_prefixes", "Number of Suppressed Internal Prefixes in the RIB.", bgpRIBLabels),
		"RIBPendingPrefixCount":            colPromDesc(bgpSubsystem, "rib_pending_prefixes", "Number of Pending Prefixes in the RIB.", bgpRIBLabels),
		"PeerTypesUp":                      colPromDesc(bgpSubsystem, "peer_types_up", "Total Number of Peer Types that are Up.", bgpPeerTypeLabels),
	}

	bgpErrors      = []error{}
	totalBGPErrors = 0.0
)

// BGPCollector collects BGP metrics, implemented as per the Collector bgp.
type BGPCollector struct{}

// NewBGPCollector returns a BGPCollector type.
func NewBGPCollector() *BGPCollector {
	return &BGPCollector{}
}

// Name of the collector. Used to parse the configuration file.
func (*BGPCollector) Name() string {
	return bgpSubsystem
}

// CollectErrors returns what errors have been gathered.
func (*BGPCollector) CollectErrors() []error {
	errors := bgpErrors
	bgpErrors = []error{}
	return errors
}

// CollectTotalErrors collects total errors.
func (*BGPCollector) CollectTotalErrors() float64 {
	return totalBGPErrors
}

// Describe all metrics
func (*BGPCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range bgpDesc {
		ch <- desc
	}
}

// Collect metric from a passed netconf.Session.
func (c *BGPCollector) Collect(ch chan<- prometheus.Metric) {
	s, err := netconf.DialSSH(sshTarget, sshClientConfig)
	if err != nil {
		totalBGPErrors++
		bgpErrors = append(bgpErrors, fmt.Errorf("could not connect to %q: %s", sshTarget, err))
		return
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-bgp-summary-information/>`))
	if err != nil {
		totalBGPErrors++
		bgpErrors = append(bgpErrors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return
	}
	if err := processBGPNetconfReply(reply, ch); err != nil {
		totalBGPErrors++
		bgpErrors = append(bgpErrors, err)
	}
}

func processBGPNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
	var netconfReply bgpRPCReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	peerTypes := make(map[string]float64)
	for _, peerData := range netconfReply.BGPInformation.BGPPeer {
		peerLabels := []string{strings.TrimSpace(peerData.PeerAddress.Text)}
		if strings.ToLower(peerData.PeerState.Text) == "established" {
			ch <- prometheus.MustNewConstMetric(bgpDesc["PeerPeerState"], prometheus.GaugeValue, 1.0, peerLabels...)
			if peerData.Description.Text != "" {
				var peerType bgpPeerType
				if err := json.Unmarshal([]byte(peerData.Description.Text), &peerType); err != nil {
					goto NoPeerType
				}
				if peerType.Type != "" {
					if _, exists := peerTypes[strings.TrimSpace(peerType.Type)]; exists {
						peerTypes[strings.TrimSpace(peerType.Type)]++
					} else {
						peerTypes[strings.TrimSpace(peerType.Type)] = 1
					}
				}
			}
		NoPeerType:
		} else {
			ch <- prometheus.MustNewConstMetric(bgpDesc["PeerPeerState"], prometheus.GaugeValue, 0.0, peerLabels...)
		}
		newCounter(ch, bgpDesc["PeerInputMessages"], peerData.InputMessages.Text, peerLabels...)
		newCounter(ch, bgpDesc["PeerOutputMessages"], peerData.OutputMessages.Text, peerLabels...)
		newGauge(ch, bgpDesc["PeerRouteQueueCount"], peerData.RouteQueueCount.Text, peerLabels...)
		newCounter(ch, bgpDesc["PeerFlapCount"], peerData.FlapCount.Text, peerLabels...)
		newGauge(ch, bgpDesc["PeerElapsedTime"], peerData.ElapsedTime.Seconds, peerLabels...)
		for _, ribData := range peerData.BGPRIB {
			peerRIBLabels := append(peerLabels, ribData.Name.Text)
			newGauge(ch, bgpDesc["PeerRIBActivePrefixCount"], ribData.ActivePrefixCount.Text, peerRIBLabels...)
			newGauge(ch, bgpDesc["PeerRIBReceivedPrefixCount"], ribData.ReceivedPrefixCount.Text, peerRIBLabels...)
			newGauge(ch, bgpDesc["PeerRIBAcceptedPrefixCount"], ribData.AcceptedPrefixCount.Text, peerRIBLabels...)
			newGauge(ch, bgpDesc["PeerRIBSuppressedPrefixCount"], ribData.SuppressedPrefixCount.Text, peerRIBLabels...)
		}

	}
	for peerType, count := range peerTypes {
		ch <- prometheus.MustNewConstMetric(bgpDesc["PeerTypesUp"], prometheus.GaugeValue, count, peerType)
	}

	newGauge(ch, bgpDesc["GroupCount"], netconfReply.BGPInformation.GroupCount.Text)
	newGauge(ch, bgpDesc["PeerCount"], netconfReply.BGPInformation.PeerCount.Text)
	newGauge(ch, bgpDesc["DownPeerCount"], netconfReply.BGPInformation.DownPeerCount.Text)

	for _, ribData := range netconfReply.BGPInformation.BGPRIB {
		ribLabels := []string{ribData.Name.Text}
		newGauge(ch, bgpDesc["RIBTotalPrefixCount"], ribData.TotalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBReceivedPrefixCount"], ribData.ReceivedPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBAcceptedPrefixCount"], ribData.AcceptedPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBActivePrefixCount"], ribData.ActivePrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBSuppressedPrefixCount"], ribData.SuppressedPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBHistoryPrefixCount"], ribData.HistoryPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBDampedPrefixCount"], ribData.DampedPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBTotalExternalPrefixCount"], ribData.TotalExternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBActiveExternalPrefixCount"], ribData.ActiveExternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBAcceptedExternalPrefixCount"], ribData.AcceptedExternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBSuppressedExternalPrefixCount"], ribData.SuppressedExternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBTotalInternalPrefixCount"], ribData.TotalInternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBActiveInternalPrefixCount"], ribData.ActiveInternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBAcceptedInternalPrefixCount"], ribData.AcceptedInternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBSuppressedInternalPrefixCount"], ribData.SuppressedInternalPrefixCount.Text, ribLabels...)
		newGauge(ch, bgpDesc["RIBPendingPrefixCount"], ribData.PendingPrefixCount.Text, ribLabels...)
	}

	return nil
}

type bgpRPCReply struct {
	XMLName        xml.Name       `xml:"rpc-reply"`
	BGPInformation bgpInformation `xml:"bgp-information"`
}

type bgpInformation struct {
	GroupCount    bgpText   `xml:"group-count"`
	PeerCount     bgpText   `xml:"peer-count"`
	DownPeerCount bgpText   `xml:"down-peer-count"`
	BGPRIB        []bgpRIB  `xml:"bgp-rib"`
	BGPPeer       []bgpPeer `xml:"bgp-peer"`
}

type bgpText struct {
	Text string `xml:",chardata"`
}

type bgpSeconds struct {
	Seconds string `xml:"seconds,attr"`
}

type bgpRIB struct {
	Name                          bgpText `xml:"name"`
	TotalPrefixCount              bgpText `xml:"total-prefix-count"`
	ReceivedPrefixCount           bgpText `xml:"received-prefix-count"`
	AcceptedPrefixCount           bgpText `xml:"accepted-prefix-count"`
	ActivePrefixCount             bgpText `xml:"active-prefix-count"`
	SuppressedPrefixCount         bgpText `xml:"suppressed-prefix-count"`
	HistoryPrefixCount            bgpText `xml:"history-prefix-count"`
	DampedPrefixCount             bgpText `xml:"damped-prefix-count"`
	TotalExternalPrefixCount      bgpText `xml:"total-external-prefix-count"`
	ActiveExternalPrefixCount     bgpText `xml:"active-external-prefix-count"`
	AcceptedExternalPrefixCount   bgpText `xml:"accepted-external-prefix-count"`
	SuppressedExternalPrefixCount bgpText `xml:"suppressed-external-prefix-count"`
	TotalInternalPrefixCount      bgpText `xml:"total-internal-prefix-count"`
	ActiveInternalPrefixCount     bgpText `xml:"active-internal-prefix-count"`
	AcceptedInternalPrefixCount   bgpText `xml:"accepted-internal-prefix-count"`
	SuppressedInternalPrefixCount bgpText `xml:"suppressed-internal-prefix-count"`
	PendingPrefixCount            bgpText `xml:"pending-prefix-count"`
	BGPRIBState                   bgpText `xml:"bgp-rib-state"`
}

type bgpPeer struct {
	PeerAddress     bgpText      `xml:"peer-address"`
	PeerAs          bgpText      `xml:"peer-as"`
	InputMessages   bgpText      `xml:"input-messages"`
	OutputMessages  bgpText      `xml:"output-messages"`
	RouteQueueCount bgpText      `xml:"route-queue-count"`
	FlapCount       bgpText      `xml:"flap-count"`
	Description     bgpText      `xml:"description"`
	ElapsedTime     bgpSeconds   `xml:"elapsed-time"`
	PeerState       bgpText      `xml:"peer-state"`
	BGPRIB          []bgpPeerRIB `xml:"bgp-rib"`
}

type bgpPeerRIB struct {
	Name                  bgpText `xml:"name"`
	ActivePrefixCount     bgpText `xml:"active-prefix-count"`
	ReceivedPrefixCount   bgpText `xml:"received-prefix-count"`
	AcceptedPrefixCount   bgpText `xml:"accepted-prefix-count"`
	SuppressedPrefixCount bgpText `xml:"suppressed-prefix-count"`
}

type bgpPeerType struct {
	Type string `json:"type"`
}
