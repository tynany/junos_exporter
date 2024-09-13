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
	ipsecSubsystem   = "ipsec"
	totalIpsecErrors = 0.0

	ipsecLabels = map[string][]string{"TunnelInformation": []string{"saremotegateway", "satunnelindex"}}
	ipsecDesc   = map[string]*prometheus.Desc{
		"TunnelStatusUp": colPromDesc(ipsecSubsystem, "tunnel_status_up", "Tunnel Status (1 UP, 0 DOWN)", ipsecLabels["TunnelInformation"]),
	}
)

// EnvCollector collects environment metrics, implemented as per the Collector interface.
type IpsecCollector struct {
	logger log.Logger
}

// NewEnvCollector returns a new EnvCollector.
func NewIpsecCollector(logger log.Logger) *IpsecCollector {
	return &IpsecCollector{logger: logger}
}

// Name of the collector.
func (*IpsecCollector) Name() string {
	return ipsecSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *IpsecCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}
	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalIpsecErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalIpsecErrors
	}
	defer s.Close()

	// IPsec inactive tunnels
	reply, err := s.Exec(netconf.RawMethod(`<get-inactive-tunnels/>`))
	if err != nil {
		totalIpsecErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalIpsecErrors
	}

	if err := processIpsecInactiveNetconfReply(reply, ch, conf.SSHTarget); err != nil {
		totalIpsecErrors++
		errors = append(errors, err)
	}

	// IPsec active tunnels
	reply, err = s.Exec(netconf.RawMethod(`<get-security-associations-information/>`))
	if err != nil {
		totalIpsecErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalIpsecErrors
	}

	if err := processIpsecActiveNetconfReply(reply, ch); err != nil {
		totalIpsecErrors++
		errors = append(errors, err)
	}

	return errors, totalIpsecErrors
}

func processIpsecInactiveNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric, sshTarget string) error {
	var netconfInactiveTunnelReply InactiveTunnelReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfInactiveTunnelReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf inactive tunnel reply xml: %s", err)
	}

	// Send tunnel status of inactive tunnels to Prometheus channel.
	// Set tunnel_status_up to 0
	for _, ipsecData := range netconfInactiveTunnelReply.IpsecUnestablishedTunnelInformation.IpsecSecurityAssociationsBlock {
		saRemoteGateway := strings.Trim(ipsecData.IpsecSecurityAssociations.SaRemoteGateway, "\n")
		saTunnelIndex := strings.Trim(ipsecData.IpsecSecurityAssociations.SaTunnelIndex, "\n")
		ch <- prometheus.MustNewConstMetric(ipsecDesc["TunnelStatusUp"], prometheus.GaugeValue, 0, saRemoteGateway, saTunnelIndex)
	}

	return nil
}

// Process active ipsec SAs.
func processIpsecActiveNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric) error {
	var netconfActiveTunnelReply ActiveTunnelReply

	if err := xml.Unmarshal([]byte(reply.RawReply), &netconfActiveTunnelReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf active tunnel reply xml: %s", err)
	}

	// Send tunnel status of active tunnels to Prometheus channel.
	// Set tunnel_status_up to 1
	for _, ipsecData := range netconfActiveTunnelReply.IpsecSecurityAssociationsInformation.IpsecSecurityAssociationsBlock {
		saRemoteGateway := strings.Trim(ipsecData.IpsecSecurityAssociations.SaRemoteGateway, "\n")
		saTunnelIndex := strings.Trim(ipsecData.IpsecSecurityAssociations.SaTunnelIndex, "\n")
		ch <- prometheus.MustNewConstMetric(ipsecDesc["TunnelStatusUp"], prometheus.GaugeValue, 1, saRemoteGateway, saTunnelIndex)
	}

	return nil
}

type InactiveTunnelReply struct {
	XMLName                             xml.Name `xml:"rpc-reply"`
	Text                                string   `xml:",chardata"`
	Junos                               string   `xml:"junos,attr"`
	MessageID                           string   `xml:"message-id,attr"`
	Xmlns                               string   `xml:"xmlns,attr"`
	IpsecUnestablishedTunnelInformation struct {
		Text                                         string `xml:",chardata"`
		Style                                        string `xml:"style,attr"`
		TotalInactiveTunnels                         string `xml:"total-inactive-tunnels"`
		TotalInactiveTunnelsWithEstablishImmediately string `xml:"total-inactive-tunnels-with-establish-immediately"`
		IpsecSecurityAssociationsBlock               []struct {
			Text                      string `xml:",chardata"`
			IpsecSecurityAssociations struct {
				Text                  string `xml:",chardata"`
				SaTunnelIndex         string `xml:"sa-tunnel-index"`
				SaRemoteGateway       string `xml:"sa-remote-gateway"`
				SaPort                string `xml:"sa-port"`
				SaTunnelEventTime     string `xml:"sa-tunnel-event-time"`
				SaTunnelEvent         string `xml:"sa-tunnel-event"`
				SaTunnelEventNumTimes string `xml:"sa-tunnel-event-num-times"`
			} `xml:"ipsec-security-associations"`
		} `xml:"ipsec-security-associations-block"`
	} `xml:"ipsec-unestablished-tunnel-information"`
}

type ActiveTunnelReply struct {
	XMLName                              xml.Name `xml:"rpc-reply"`
	Text                                 string   `xml:",chardata"`
	Junos                                string   `xml:"junos,attr"`
	IpsecSecurityAssociationsInformation struct {
		Text                           string `xml:",chardata"`
		Style                          string `xml:"style,attr"`
		TotalActiveTunnels             string `xml:"total-active-tunnels"`
		IpsecSecurityAssociationsBlock []struct {
			Text                      string `xml:",chardata"`
			SaBlockState              string `xml:"sa-block-state"`
			IpsecSecurityAssociations struct {
				Text                     string `xml:",chardata"`
				SaDirection              string `xml:"sa-direction"`
				SaTunnelIndex            string `xml:"sa-tunnel-index"`
				SaSpi                    string `xml:"sa-spi"`
				SaAuxSpi                 string `xml:"sa-aux-spi"`
				SaRemoteGateway          string `xml:"sa-remote-gateway"`
				SaPort                   string `xml:"sa-port"`
				SaVpnMonitoringState     string `xml:"sa-vpn-monitoring-state"`
				SaProtocol               string `xml:"sa-protocol"`
				SaEspEncryptionAlgorithm string `xml:"sa-esp-encryption-algorithm"`
				SaHmacAlgorithm          string `xml:"sa-hmac-algorithm"`
				SaHardLifetime           string `xml:"sa-hard-lifetime"`
				SaLifesizeRemaining      string `xml:"sa-lifesize-remaining"`
				SaVirtualSystem          string `xml:"sa-virtual-system"`
			} `xml:"ipsec-security-associations"`
		} `xml:"ipsec-security-associations-block"`
	} `xml:"ipsec-security-associations-information"`
	Cli struct {
		Text   string `xml:",chardata"`
		Banner string `xml:"banner"`
	} `xml:"cli"`
}
