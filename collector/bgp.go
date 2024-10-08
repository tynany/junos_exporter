package collector

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	bgpSubsystem = "bgp"

	bgpPeerLabels     = []string{"peer", "interface", "peer_address_family"}
	bgpRIBLabels      = []string{"routing_instance"}
	bgpPeerRIBLabels  = append(bgpPeerLabels, bgpRIBLabels...)
	bgpPeerTypeLabels = []string{"type"}
	bgpDesc           = map[string]*prometheus.Desc{
		"GroupCount":                       colPromDesc(bgpSubsystem, "groups", "Number of Configured Groups.", bgpRIBLabels),
		"PeerCount":                        colPromDesc(bgpSubsystem, "peers", "Number of Configured Peers.", bgpRIBLabels),
		"DownPeerCount":                    colPromDesc(bgpSubsystem, "down_peers", "Number of Peers that are Down.", bgpRIBLabels),
		"PeerInputMessages":                colPromDesc(bgpSubsystem, "peer_input_messages", "Number of Input Messages for a Peer.", bgpPeerRIBLabels),
		"PeerOutputMessages":               colPromDesc(bgpSubsystem, "peer_output_messages", "Number of Output Messages for a Peer.", bgpPeerRIBLabels),
		"PeerRouteQueueCount":              colPromDesc(bgpSubsystem, "peer_route_queue", "Number of Route Queues for a Peer.", bgpPeerRIBLabels),
		"PeerFlapCount":                    colPromDesc(bgpSubsystem, "peer_flaps", "Number of Time the Peer has Flapped.", bgpPeerRIBLabels),
		"PeerElapsedTime":                  colPromDesc(bgpSubsystem, "peer_elapsed_time_seconds", "Length of Time the Peer has Been Up.", bgpPeerRIBLabels),
		"PeerPeerState":                    colPromDesc(bgpSubsystem, "peer_up", "State of the Peer. (1 = Established, 0 = Down).", bgpPeerRIBLabels),
		"PeerRIBActivePrefixCount":         colPromDesc(bgpSubsystem, "peer_rib_active_prefixes", "Number of Active Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBReceivedPrefixCount":       colPromDesc(bgpSubsystem, "peer_rib_received_prefixes", "Number of Received Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBAcceptedPrefixCount":       colPromDesc(bgpSubsystem, "peer_rib_accepted_prefixes", "Number of Accepted Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBSuppressedPrefixCount":     colPromDesc(bgpSubsystem, "peer_rib_suppressed_prefixes", "Number of Suppressed Prefixes for the Peer.", bgpPeerRIBLabels),
		"PeerRIBAdvertisedPrefixCount":     colPromDesc(bgpSubsystem, "peer_rib_advertised_prefixes", "Number of Advertised Prefixes for the Peer.", bgpPeerRIBLabels),
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
	totalBGPErrors = 0.0
)

// BGPCollector collects BGP metrics, implemented as per the Collector interface.
type BGPCollector struct {
	logger log.Logger
}

// NewBGPCollector returns a new BGPCollector
func NewBGPCollector(logger log.Logger) *BGPCollector {
	return &BGPCollector{logger: logger}
}

// Name of the collector.
func (*BGPCollector) Name() string {
	return bgpSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *BGPCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalBGPErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalBGPErrors
	}
	defer s.Close()

	// show bgp summary | display xml
	reply, err := s.Exec(netconf.RawMethod(`<get-bgp-summary-information/>`))
	if err != nil {
		totalBGPErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalBGPErrors
	}

	// show bgp neighbor | display xml
	replyNeighbor, err := s.Exec(netconf.RawMethod(`<get-bgp-neighbor-information/>`))
	if err != nil {
		totalBGPErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalBGPErrors
	}

	// show route instance | display xml
	replyRouteInstance, err := s.Exec(netconf.RawMethod(`<get-instance-information/>`))
	if err != nil {
		totalBGPErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalBGPErrors
	}

	bgpPeerInterfaces, err := getBgpPeerInterface(replyNeighbor)
	if err != nil {
		totalBGPErrors++
		errors = append(errors, err)
	}

	routeInstances, routeInstanceNames, err := getInstanceNameToRibName(replyRouteInstance)
	if err != nil {
		totalBGPErrors++
		errors = append(errors, err)
	}

	replyBgpSummaryInstance := make(map[string]*netconf.RPCReply)
	for routeInstance := range routeInstanceNames {
		// filter out default routing instances that are reserved so we dont pull un-necessary RPC calls
		if !strings.Contains(routeInstance, "__master") && !strings.Contains(routeInstance, "__juniper") && !strings.Contains(routeInstance, "mgmt_junos") {
			routeInstanceCommand := fmt.Sprintf(`<get-bgp-summary-information><instance>%s</instance></get-bgp-summary-information>`, routeInstance)
			replyBgpSummaryVrf, err := s.Exec(netconf.RawMethod(routeInstanceCommand))
			replyBgpSummaryInstance[routeInstance] = replyBgpSummaryVrf
			if err != nil {
				totalBGPErrors++
				errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
				return errors, totalBGPErrors
			}
		}
	}

	if err := processBGPNetconfReply(
		reply,
		replyNeighbor,
		replyBgpSummaryInstance,
		ch,
		conf.BGPTypeKeys,
		bgpPeerInterfaces,
		routeInstances,
		routeInstanceNames,
		c.logger,
	); err != nil {
		totalBGPErrors++
		errors = append(errors, err)
	}

	if err := processBGPNeighborNetconfReply(replyNeighbor, ch, c.logger); err != nil {
		totalBGPErrors++
		errors = append(errors, err)
	}
	return errors, totalBGPErrors
}

func getInstanceNameToRibName(reply *netconf.RPCReply) (map[string]string, map[string]string, error) {
	instanceToIribName := make(map[string]string)
	routeInstanceNames := make(map[string]string)
	var netconfRouteInstanceReply routeInstanceRPCReply
	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfRouteInstanceReply); err != nil {
		return instanceToIribName, routeInstanceNames, fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, instanceCore := range netconfRouteInstanceReply.InstanceInformation.InstanceCore {
		routeInstanceNames[instanceCore.InstanceName] = instanceCore.InstanceName
		for _, iriB := range instanceCore.InstanceRib {
			instanceToIribName[iriB.IribName] = instanceCore.InstanceName
		}
	}
	return instanceToIribName, routeInstanceNames, nil
}

func getBgpPeerInterface(reply *netconf.RPCReply) (map[string]string, error) {
	peerToInterface := make(map[string]string)
	var netconfReply bgpNeighborRPCReply
	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return peerToInterface, fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, peerData := range netconfReply.BgpInformation.BgpPeer {
		if peerData.PeerAddress != "" {
			// Junos 17 & 18 uses <peer-address>
			peerAddSplit := strings.Split(peerData.PeerAddress, "+")
			peerToInterface[peerAddSplit[0]] = peerData.LocalInterfaceName
		} else if peerData.BGPPeerHeader.PeerAddress != "" {
			// Junos 19 uses <bgp-peer-header><peer-address>
			peerAddSplit := strings.Split(peerData.BGPPeerHeader.PeerAddress, "+")
			peerToInterface[peerAddSplit[0]] = peerData.LocalInterfaceName
		}
	}
	return peerToInterface, nil
}

func processBGPNeighborNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric, logger log.Logger) error {
	var netconfReply bgpNeighborRPCReply
	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}
	for _, peerData := range netconfReply.BgpInformation.BgpPeer {
		// define <peer-address>
		peerAddress := ""
		if peerData.PeerAddress != "" {
			// Junos 17 & 18 uses <peer-address>
			peerAddress = peerData.PeerAddress
		} else if peerData.BGPPeerHeader.PeerAddress != "" {
			// Junos 19 uses <bgp-peer-header><peer-address>
			peerAddress = peerData.BGPPeerHeader.PeerAddress
		}

		// define <local-interface-name>
		localInterfaceName := peerData.LocalInterfaceName

		// define peer labels
		if peerAddress != "" {
			re := regexp.MustCompile(`\+.*`)
			peerLabels := []string{
				re.ReplaceAllString(strings.TrimSpace(peerAddress), ""),
				strings.TrimSpace(localInterfaceName),
				strings.TrimSpace(peerData.PeerCfgRti.Text),
				strings.TrimSpace(peerData.NlriTypePeer),
			}
			newGauge(logger, ch, bgpDesc["PeerRIBAdvertisedPrefixCount"], peerData.BGPRIB.AdvertisedPrefixCount, peerLabels...)
		}
	}
	return nil
}

func processBGPNetconfReply(
	reply *netconf.RPCReply,
	replyNeighbor *netconf.RPCReply,
	replyBgpSummaryInstance map[string]*netconf.RPCReply,
	ch chan<- prometheus.Metric,
	bgpTypeKeys []string,
	bgpPeerInterfaces map[string]string,
	routeInstances map[string]string,
	routeInstanceNames map[string]string,
	logger log.Logger,
) error {

	var netconfReply bgpNeighborRPCReply
	var netconfInfoReply bgpRPCReply
	routeInstanceCheck := make(map[string]string)

	if err := xml.Unmarshal([]byte(replyNeighbor.RawReply), &netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfInfoReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}

	// loop through route instances, unmarshall, only gather RIB totals once per route instance
	for routeInstanceBGP := range replyBgpSummaryInstance {
		var netconfReplyInstance bgpRPCReplyInstance
		if err := xml.Unmarshal([]byte(replyBgpSummaryInstance[routeInstanceBGP].RawReply), &netconfReplyInstance); err != nil {
			return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
		}
		ribLabels := []string{routeInstanceBGP}
		newGauge(logger, ch, bgpDesc["GroupCount"], netconfReplyInstance.BgpInformation.GroupCount, ribLabels...)
		newGauge(logger, ch, bgpDesc["PeerCount"], netconfReplyInstance.BgpInformation.PeerCount, ribLabels...)
		newGauge(logger, ch, bgpDesc["DownPeerCount"], netconfReplyInstance.BgpInformation.DownPeerCount, ribLabels...)
		for routeInstance := range routeInstances {
			if routeInstanceBGP == routeInstances[routeInstance] {
				for _, ribData := range netconfReplyInstance.BgpInformation.BgpRib {
					if ribData.Name == routeInstance {
						if val, ok := routeInstanceCheck[routeInstanceBGP]; !ok {
							routeInstanceCheck[routeInstanceBGP] = val
							newGauge(logger, ch, bgpDesc["RIBTotalPrefixCount"], ribData.TotalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBHistoryPrefixCount"], ribData.HistoryPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBDampedPrefixCount"], ribData.DampedPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBTotalExternalPrefixCount"], ribData.TotalExternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBActiveExternalPrefixCount"], ribData.ActiveExternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBAcceptedExternalPrefixCount"], ribData.AcceptedExternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBSuppressedExternalPrefixCount"], ribData.SuppressedExternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBTotalInternalPrefixCount"], ribData.TotalInternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBActiveInternalPrefixCount"], ribData.ActiveInternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBAcceptedInternalPrefixCount"], ribData.AcceptedInternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBSuppressedInternalPrefixCount"], ribData.SuppressedInternalPrefixCount, ribLabels...)
							newGauge(logger, ch, bgpDesc["RIBPendingPrefixCount"], ribData.PendingPrefixCount, ribLabels...)
						}
					}
				}
			}
		}
	}

	peerTypes := make(map[string]float64)

	for _, peerData := range netconfReply.BgpInformation.BgpPeer {
		peerInterfaceLocal := ""
		peerAddress := ""
		peerAddressFamily := peerData.NlriTypePeer

		if peerData.PeerAddress != "" {
			// Junos 17 & 18 uses <peer-address>
			peerAddressSplit := strings.Split(peerData.PeerAddress, "+")
			peerAddress = peerAddressSplit[0]
		} else if peerData.BGPPeerHeader.PeerAddress != "" {
			// Junos 19 uses <bgp-peer-header><peer-address>
			peerAddressSplit := strings.Split(peerData.BGPPeerHeader.PeerAddress, "+")
			peerAddress = peerAddressSplit[0]
		}

		if val, exists := bgpPeerInterfaces[peerAddress]; exists {
			peerInterfaceLocal = val
		}

		peerLabels := []string{strings.TrimSpace(peerAddress), strings.TrimSpace(peerInterfaceLocal), strings.TrimSpace(peerAddressFamily)}
		var peerType map[string]string

		if len(bgpTypeKeys) > 0 && peerData.Description.Text != "" {
			if err := json.Unmarshal([]byte(peerData.Description.Text), &peerType); err == nil {
				for _, descKey := range bgpTypeKeys {
					if peerType[descKey] != "" {
						if _, exist := peerTypes[strings.TrimSpace(peerType[descKey])]; !exist {
							peerTypes[strings.TrimSpace(peerType[descKey])] = 0
						}
					}
				}
			}
		}

		peerRIBLabels := append(peerLabels, peerData.PeerCfgRti.Text)

		if strings.ToLower(peerData.PeerState.Text) == "established" {
			ch <- prometheus.MustNewConstMetric(bgpDesc["PeerPeerState"], prometheus.GaugeValue, 1.0, peerRIBLabels...)
			for _, descKey := range bgpTypeKeys {
				if peerType[descKey] != "" {
					peerTypes[strings.TrimSpace(peerType[descKey])]++
				}
			}
		} else {
			ch <- prometheus.MustNewConstMetric(bgpDesc["PeerPeerState"], prometheus.GaugeValue, 0.0, peerRIBLabels...)
		}

		newCounter(logger, ch, bgpDesc["PeerInputMessages"], peerData.InputMessages.Text, peerRIBLabels...)
		newCounter(logger, ch, bgpDesc["PeerOutputMessages"], peerData.OutputMessages.Text, peerRIBLabels...)
		newGauge(logger, ch, bgpDesc["PeerRouteQueueCount"], peerData.RouteQueueCount.Text, peerRIBLabels...)
		newCounter(logger, ch, bgpDesc["PeerFlapCount"], peerData.FlapCount.Text, peerRIBLabels...)
		newGauge(logger, ch, bgpDesc["PeerElapsedTime"], peerData.ElapsedTime.Seconds, peerRIBLabels...)
		newGauge(logger, ch, bgpDesc["PeerRIBActivePrefixCount"], peerData.BGPRIB.ActivePrefixCount, peerRIBLabels...)
		newGauge(logger, ch, bgpDesc["PeerRIBReceivedPrefixCount"], peerData.BGPRIB.ReceivedPrefixCount, peerRIBLabels...)
		newGauge(logger, ch, bgpDesc["PeerRIBAcceptedPrefixCount"], peerData.BGPRIB.AcceptedPrefixCount, peerRIBLabels...)
		newGauge(logger, ch, bgpDesc["PeerRIBSuppressedPrefixCount"], peerData.BGPRIB.SuppressedPrefixCount, peerRIBLabels...)
	}

	for peerType, count := range peerTypes {
		ch <- prometheus.MustNewConstMetric(bgpDesc["PeerTypesUp"], prometheus.GaugeValue, count, peerType)
	}

	return nil
}

// ********************* show route instance START ********************* //
type routeInstanceRPCReply struct {
	XMLName             xml.Name                 `xml:"rpc-reply"`
	InstanceInformation routeInstanceInformation `xml:"instance-information"`
}

type routeInstanceInformation struct {
	InstanceCore []routeInstanceCore `xml:"instance-core"`
}

type routeInstanceCore struct {
	InstanceName string             `xml:"instance-name"`
	InstanceType string             `xml:"instance-type"`
	InstanceRib  []routeInstanceRib `xml:"instance-rib"`
}

type routeInstanceRib struct {
	Text              string `xml:",chardata"`
	IribName          string `xml:"irib-name"`
	IribActiveCount   string `xml:"irib-active-count"`
	IribHolddownCount string `xml:"irib-holddown-count"`
	IribHiddenCount   string `xml:"irib-hidden-count"`
}

// ********************* show route instance END *********************

// ********************* show bgp summary instance [INSTANCE] START ********************* //
type bgpRPCReplyInstance struct {
	XMLName        xml.Name           `xml:"rpc-reply"`
	BgpInformation bgpSummaryInstance `xml:"bgp-information"`
}
type bgpSummaryInstance struct {
	Text          string                   `xml:",chardata"`
	Xmlns         string                   `xml:"xmlns,attr"`
	BgpThreadMode string                   `xml:"bgp-thread-mode"`
	ThreadState   string                   `xml:"thread-state"`
	GroupCount    string                   `xml:"group-count"`
	PeerCount     string                   `xml:"peer-count"`
	DownPeerCount string                   `xml:"down-peer-count"`
	BgpRib        []bgpSummaryInstanceRib  `xml:"bgp-rib"`
	BgpPeer       []bgpSummaryInstancePeer `xml:"bgp-peer"`
}
type bgpSummaryInstanceRib struct {
	Text                          string `xml:",chardata"`
	Name                          string `xml:"name"`
	TotalPrefixCount              string `xml:"total-prefix-count"`
	ReceivedPrefixCount           string `xml:"received-prefix-count"`
	AcceptedPrefixCount           string `xml:"accepted-prefix-count"`
	ActivePrefixCount             string `xml:"active-prefix-count"`
	SuppressedPrefixCount         string `xml:"suppressed-prefix-count"`
	HistoryPrefixCount            string `xml:"history-prefix-count"`
	DampedPrefixCount             string `xml:"damped-prefix-count"`
	TotalExternalPrefixCount      string `xml:"total-external-prefix-count"`
	ActiveExternalPrefixCount     string `xml:"active-external-prefix-count"`
	AcceptedExternalPrefixCount   string `xml:"accepted-external-prefix-count"`
	SuppressedExternalPrefixCount string `xml:"suppressed-external-prefix-count"`
	TotalInternalPrefixCount      string `xml:"total-internal-prefix-count"`
	ActiveInternalPrefixCount     string `xml:"active-internal-prefix-count"`
	AcceptedInternalPrefixCount   string `xml:"accepted-internal-prefix-count"`
	SuppressedInternalPrefixCount string `xml:"suppressed-internal-prefix-count"`
	PendingPrefixCount            string `xml:"pending-prefix-count"`
	BgpRibState                   string `xml:"bgp-rib-state"`
	VpnRibState                   string `xml:"vpn-rib-state"`
}
type bgpSummaryInstancePeer struct {
	Text            string                            `xml:",chardata"`
	PeerAddress     string                            `xml:"peer-address"`
	PeerAs          string                            `xml:"peer-as"`
	InputMessages   string                            `xml:"input-messages"`
	OutputMessages  string                            `xml:"output-messages"`
	RouteQueueCount string                            `xml:"route-queue-count"`
	FlapCount       string                            `xml:"flap-count"`
	ElapsedTime     bgpSummaryInstancePeerElapsedTime `xml:"elapsed-time"`
	PeerState       bgpSummaryInstancePeerPeerState   `xml:"peer-state"`
	BgpRib          bgpSummaryInstancePeerBgpRib      `xml:"bgp-rib"`
	Description     string                            `xml:"description"`
}
type bgpSummaryInstancePeerElapsedTime struct {
	Text    string `xml:",chardata"`
	Seconds string `xml:"seconds,attr"`
}
type bgpSummaryInstancePeerPeerState struct {
	Text   string `xml:",chardata"`
	Format string `xml:"format,attr"`
}
type bgpSummaryInstancePeerBgpRib struct {
	Text                  string `xml:",chardata"`
	Name                  string `xml:"name"`
	ActivePrefixCount     string `xml:"active-prefix-count"`
	ReceivedPrefixCount   string `xml:"received-prefix-count"`
	AcceptedPrefixCount   string `xml:"accepted-prefix-count"`
	SuppressedPrefixCount string `xml:"suppressed-prefix-count"`
}

// ********************* show bgp summary instance [INSTANCE] END ********************* //

// ********************* show bgp neighbor START ********************* //
type bgpNeighborRPCReply struct {
	XMLName        xml.Name               `xml:"rpc-reply"`
	BgpInformation bgpNeighborInformation `xml:"bgp-information"`
}
type bgpNeighborInformation struct {
	BgpPeer []bgpNeighborPeer `xml:"bgp-peer"`
}
type bgpNeighborPeer struct {
	PeerAddress        string        `xml:"peer-address"`
	BGPPeerHeader      bgpPeerHeader `xml:"bgp-peer-header"`
	LocalInterfaceName string        `xml:"local-interface-name"`
	PeerAs             bgpText       `xml:"peer-as"`
	InputMessages      bgpText       `xml:"input-messages"`
	OutputMessages     bgpText       `xml:"output-messages"`
	RouteQueueCount    bgpText       `xml:"route-queue-count"`
	FlapCount          bgpText       `xml:"flap-count"`
	Description        bgpText       `xml:"description"`
	ElapsedTime        bgpSeconds    `xml:"elapsed-time"`
	PeerState          bgpText       `xml:"peer-state"`
	BGPRIB             bgpPeerRIB    `xml:"bgp-rib"`
	PeerCfgRti         bgpText       `xml:"peer-cfg-rti"`
	NlriTypePeer       string        `xml:"nlri-type-peer"`
}
type bgpPeerRIB struct {
	Name                  string `xml:"name"`
	RibBit                string `xml:"rib-bit"`
	BgpRibState           string `xml:"bgp-rib-state"`
	VpnRibState           string `xml:"vpn-rib-state"`
	SendState             string `xml:"send-state"`
	ActivePrefixCount     string `xml:"active-prefix-count"`
	ReceivedPrefixCount   string `xml:"received-prefix-count"`
	AcceptedPrefixCount   string `xml:"accepted-prefix-count"`
	SuppressedPrefixCount string `xml:"suppressed-prefix-count"`
	AdvertisedPrefixCount string `xml:"advertised-prefix-count"`
}
type bgpPeerHeader struct {
	PeerAddress  string `xml:"peer-address"`
	PeerAs       string `xml:"peer-as"`
	LocalAddress string `xml:"local-address"`
	LocalAs      string `xml:"local-as"`
}

// ********************* show bgp neighbor END *********************

// ********************* show bgp summary START *********************
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
	PeerAddress        bgpText    `xml:"peer-address"`
	PeerAs             bgpText    `xml:"peer-as"`
	InputMessages      bgpText    `xml:"input-messages"`
	OutputMessages     bgpText    `xml:"output-messages"`
	RouteQueueCount    bgpText    `xml:"route-queue-count"`
	FlapCount          bgpText    `xml:"flap-count"`
	Description        bgpText    `xml:"description"`
	ElapsedTime        bgpSeconds `xml:"elapsed-time"`
	PeerState          bgpText    `xml:"peer-state"`
	BGPRIB             []bgpRIB   `xml:"bgp-rib"`
	LocalInterfaceName bgpText    `xml:"local-interface-name"`
}
type bgpText struct {
	Text string `xml:",chardata"`
}
type bgpSeconds struct {
	Seconds string `xml:"seconds,attr"`
}

// ********************* show bgp summary END *********************
