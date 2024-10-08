package collector

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ifaceSubsystem   = "interface"
	totalIfaceErrors = 0.0
)

func getInterfaceDesc(ifaceDescrKeys, ifaceMetricKeys []string) map[string]*prometheus.Desc {
	var ifacePhysicalLabels = []string{"interface"}
	var ifacePRECLClass = append(ifacePhysicalLabels, "class")

	ifaceDesc := map[string]*prometheus.Desc{
		"Up":                                       colPromDesc(ifaceSubsystem, "up", "Whether the interface is up (1 = up, 0 = down).", ifacePhysicalLabels),
		"InterfaceFlapped":                         colPromDesc(ifaceSubsystem, "interface_flapped_seconds", "How long since the last interface flap.", ifacePhysicalLabels),
		"InputBytes":                               colPromDesc(ifaceSubsystem, "input_bytes", "Input Bytes.", ifacePhysicalLabels),
		"OutputBytes":                              colPromDesc(ifaceSubsystem, "output_bytes", "Output Bytes.", ifacePhysicalLabels),
		"InputPackets":                             colPromDesc(ifaceSubsystem, "input_packets", "Input Packets.", ifacePhysicalLabels),
		"OutputPackets":                            colPromDesc(ifaceSubsystem, "output_packets", "Output Packets.", ifacePhysicalLabels),
		"InputBps":                                 colPromDesc(ifaceSubsystem, "input_bps", "Input BPS.", ifacePhysicalLabels),
		"OutputBps":                                colPromDesc(ifaceSubsystem, "output_bps", "Output BPS.", ifacePhysicalLabels),
		"InputPps":                                 colPromDesc(ifaceSubsystem, "input_pps", "Input PPS.", ifacePhysicalLabels),
		"OutputPps":                                colPromDesc(ifaceSubsystem, "output_pps", "Output PPS.", ifacePhysicalLabels),
		"V6InputBytes":                             colPromDesc(ifaceSubsystem, "ipv6_input_bytes", "Input IPv6 Bytes.", ifacePhysicalLabels),
		"V6OutputBytes":                            colPromDesc(ifaceSubsystem, "ipv6_output_bytes", "Output IPv6 Bytes.", ifacePhysicalLabels),
		"V6InputPackets":                           colPromDesc(ifaceSubsystem, "ipv6_input_packets", "Input IPv6 Packets.", ifacePhysicalLabels),
		"V6OutputPackets":                          colPromDesc(ifaceSubsystem, "ipv6_output_packets", "Output IPv6 Packets.", ifacePhysicalLabels),
		"InputErrors":                              colPromDesc(ifaceSubsystem, "input_errors", "Input Errors.", ifacePhysicalLabels),
		"InputDrops":                               colPromDesc(ifaceSubsystem, "input_drops", "Input Drops.", ifacePhysicalLabels),
		"FramingErrors":                            colPromDesc(ifaceSubsystem, "framing_errors", "Framing Errors.", ifacePhysicalLabels),
		"InputRunts":                               colPromDesc(ifaceSubsystem, "input_runts", "Input Runts.", ifacePhysicalLabels),
		"InputGiants":                              colPromDesc(ifaceSubsystem, "input_giants", "Input Giants.", ifacePhysicalLabels),
		"InputDiscards":                            colPromDesc(ifaceSubsystem, "input_discards", "Input Discards.", ifacePhysicalLabels),
		"InputResourceErrors":                      colPromDesc(ifaceSubsystem, "input_resource_errors", "Input Resource Errors.", ifacePhysicalLabels),
		"InputL3Incompletes":                       colPromDesc(ifaceSubsystem, "input_l3_incompletes", "Input L3 Incompletes.", ifacePhysicalLabels),
		"InputL2ChannelErrors":                     colPromDesc(ifaceSubsystem, "input_l2_channel_errors", "Input L2 Channel Errors.", ifacePhysicalLabels),
		"InputL2MismatchTimeouts":                  colPromDesc(ifaceSubsystem, "input_l2_mismatch_timeouts", "Input L2 Mismatch Timeouts.", ifacePhysicalLabels),
		"InputFifoErrors":                          colPromDesc(ifaceSubsystem, "input_fifo_errors", "Input FIFO Errors.", ifacePhysicalLabels),
		"CarrierTransitions":                       colPromDesc(ifaceSubsystem, "carrier_transitions", "Carrier transitions.", ifacePhysicalLabels),
		"OutputErrors":                             colPromDesc(ifaceSubsystem, "output_errors", "Output Errors.", ifacePhysicalLabels),
		"OutputDrops":                              colPromDesc(ifaceSubsystem, "output_drops", "Output Drops.", ifacePhysicalLabels),
		"MtuErrors":                                colPromDesc(ifaceSubsystem, "mtu_errors", "MTU Errors.", ifacePhysicalLabels),
		"OutputResourceErrors":                     colPromDesc(ifaceSubsystem, "output_resource_errors", "Output Resource Errors.", ifacePhysicalLabels),
		"OutputCollisions":                         colPromDesc(ifaceSubsystem, "output_collisions", "Output Collisions.", ifacePhysicalLabels),
		"AgedPackets":                              colPromDesc(ifaceSubsystem, "aged_packets", "Aged Packets.", ifacePhysicalLabels),
		"HsLinkCrcErrors":                          colPromDesc(ifaceSubsystem, "hslink_crc_errors", "HS Link CRC Errors.", ifacePhysicalLabels),
		"OutputFifoErrors":                         colPromDesc(ifaceSubsystem, "output_fifo_errors", "Output FIFO Errors.", ifacePhysicalLabels),
		"StpInputBytesDropped":                     colPromDesc(ifaceSubsystem, "stp_input_bytes_dropped", "STP Input Bytes Dropped.", ifacePhysicalLabels),
		"StpOutputBytesDropped":                    colPromDesc(ifaceSubsystem, "stp_output_bytes_dropped", "STP Output Bytes Dropped.", ifacePhysicalLabels),
		"StpInputPacketsDropped":                   colPromDesc(ifaceSubsystem, "stp_input_packets_dropped", "STP Input Packets Dropped.", ifacePhysicalLabels),
		"StpOutputPacketsDropped":                  colPromDesc(ifaceSubsystem, "stp_output_packets_dropped", "STP Input Packets Dropped", ifacePhysicalLabels),
		"BitErrorSeconds":                          colPromDesc(ifaceSubsystem, "pcs_bit_error_seconds", "The number of seconds during which at least one bit error rate (BER) occurred while the PCS receiver is operating in normal mode.", ifacePhysicalLabels),
		"ErroredBlocksSeconds":                     colPromDesc(ifaceSubsystem, "pcs_errored_blocks_seconds", "The number of seconds when at least one errored block occurred while the PCS receiver is operating in normal mode.", ifacePhysicalLabels),
		"MACInputBytes":                            colPromDesc(ifaceSubsystem, "mac_input_bytes", "MAC Input Bytes.", ifacePhysicalLabels),
		"MACOutputBytes":                           colPromDesc(ifaceSubsystem, "mac_output_bytes", "MAC Output Bytes.", ifacePhysicalLabels),
		"MACInputPackets":                          colPromDesc(ifaceSubsystem, "mac_input_packets", "MAC Input Packets.", ifacePhysicalLabels),
		"MACOutputPackets":                         colPromDesc(ifaceSubsystem, "mac_output_packets", "MAC Output Packets.", ifacePhysicalLabels),
		"MACInputUnicasts":                         colPromDesc(ifaceSubsystem, "mac_input_unicasts", "MAC Input Unicasts.", ifacePhysicalLabels),
		"MACOutputUnicasts":                        colPromDesc(ifaceSubsystem, "mac_output_unicasts", "MAC Output Unicasts.", ifacePhysicalLabels),
		"MACInputBroadcasts":                       colPromDesc(ifaceSubsystem, "mac_input_broadcasts", "MAC Input Broadcasts.", ifacePhysicalLabels),
		"MACOutputBroadcasts":                      colPromDesc(ifaceSubsystem, "mac_output_broadcasts", "MAC Output Broadcasts.", ifacePhysicalLabels),
		"MACInputMulticasts":                       colPromDesc(ifaceSubsystem, "mac_input_multicasts", "MAC Input Multicasts.", ifacePhysicalLabels),
		"MACOutputMulticasts":                      colPromDesc(ifaceSubsystem, "mac_output_multicasts", "MAC Output Multicasts.", ifacePhysicalLabels),
		"MACInputCrcErrors":                        colPromDesc(ifaceSubsystem, "mac_input_crc_errors", "MAC Input CRC Errors.", ifacePhysicalLabels),
		"MACOutputCrcErrors":                       colPromDesc(ifaceSubsystem, "mac_output_crc_errors", "MAC Output CRC Errors.", ifacePhysicalLabels),
		"MACInputFifoErrors":                       colPromDesc(ifaceSubsystem, "mac_input_fifo_errors", "MAC Input FIFO Errors.", ifacePhysicalLabels),
		"MACOutputFifoErrors":                      colPromDesc(ifaceSubsystem, "mac_output_fifo_errors", "MAC output FIFO Errors.", ifacePhysicalLabels),
		"MACInputMacControlFrames":                 colPromDesc(ifaceSubsystem, "mac_input_control_frames", "MAC Input Control Frames.", ifacePhysicalLabels),
		"MACOutputMacControlFrames":                colPromDesc(ifaceSubsystem, "mac_output_control_frames", "MAC Output Control Frames.", ifacePhysicalLabels),
		"MACInputMacPauseFrames":                   colPromDesc(ifaceSubsystem, "mac_input_pause_frames", "MAC Input Pause Frames.", ifacePhysicalLabels),
		"MACOutputMacPauseFrames":                  colPromDesc(ifaceSubsystem, "mac_output_pause_frames", "MAC Output Pause Frames.", ifacePhysicalLabels),
		"MACInputOversizedFrames":                  colPromDesc(ifaceSubsystem, "mac_input_oversized_frames", "MAC Input Oversized Frames.", ifacePhysicalLabels),
		"MACInputJabberFrames":                     colPromDesc(ifaceSubsystem, "mac_input_jabber_frames", "MAC Input Jabber Frames.", ifacePhysicalLabels),
		"MACInputFragmentFrames":                   colPromDesc(ifaceSubsystem, "mac_input_fragement_frames", "MAC Input Fragment Frames.", ifacePhysicalLabels),
		"MACInputVlanTaggedFrames":                 colPromDesc(ifaceSubsystem, "mac_input_vlan_tagged_frames", "MAC Input VLAN Tagged Frames.", ifacePhysicalLabels),
		"MACInputCodeViolations":                   colPromDesc(ifaceSubsystem, "mac_input_code_violations", "MAC Input Code Violations.", ifacePhysicalLabels),
		"MACInputTotalErrors":                      colPromDesc(ifaceSubsystem, "mac_input_errors", "MAC Input Errors.", ifacePhysicalLabels),
		"MACOutputTotalErrors":                     colPromDesc(ifaceSubsystem, "mac_output_errors", "MAC Output Errors.", ifacePhysicalLabels),
		"FilterInputPackets":                       colPromDesc(ifaceSubsystem, "filtered_input_packets", "Filtered Input Packets.", ifacePhysicalLabels),
		"FilterInputRejectCount":                   colPromDesc(ifaceSubsystem, "filtered_input_rejects", "Filtered Input Rejected.", ifacePhysicalLabels),
		"FilterInputRejectDestinationAddressCount": colPromDesc(ifaceSubsystem, "filtered_input_destination_address_rejects", "Filtered Input Reject Destinaion Address.", ifacePhysicalLabels),
		"FilterInputRejectSourceAddressCount":      colPromDesc(ifaceSubsystem, "filtered_input_source_address_rejects", "Filtered Input Reject Source Address.", ifacePhysicalLabels),
		"FilterOutputPackets":                      colPromDesc(ifaceSubsystem, "filtered_output_packets", "Filtered Output Packets.", ifacePhysicalLabels),
		"FilterOutputPacketPadCount":               colPromDesc(ifaceSubsystem, "filtered_output_packet_pads", "Filtered Output Packet Pad.", ifacePhysicalLabels),
		"FilterOutputPacketErrorCount":             colPromDesc(ifaceSubsystem, "filtered_output_packet_errors", "Filtered Output Packet Errors.", ifacePhysicalLabels),
		"FilterCamDestinationFilterCount":          colPromDesc(ifaceSubsystem, "filtered_cam_destinations", "Filtered CAM Destination.", ifacePhysicalLabels),
		"FilterCamSourceFilterCount":               colPromDesc(ifaceSubsystem, "filtered_cam_sources", "Filtered CAM Source.", ifacePhysicalLabels),
		"PreclRxPackets":                           colPromDesc(ifaceSubsystem, "precl_input_packets", "PRECL Input Packets.", ifacePRECLClass),
		"PreclTxPackets":                           colPromDesc(ifaceSubsystem, "precl_output_packets", "PRECL Output Packets.", ifacePRECLClass),
		"PreclDroppedPackets":                      colPromDesc(ifaceSubsystem, "precl_dropped_packets", "PRECL Dropped Packets.", ifacePRECLClass),
		"FecCcwCount":                              colPromDesc(ifaceSubsystem, "fec_ccw", "FEC CCW Count.", ifacePhysicalLabels),
		"FecNccwCount":                             colPromDesc(ifaceSubsystem, "fec_nccw", "FEC NCCW Count.", ifacePhysicalLabels),
		"FecCcwErrorRate":                          colPromDesc(ifaceSubsystem, "fec_ccw_error_rate", "FEC CCW Error Rate.", ifacePhysicalLabels),
		"FecNccwErrorRate":                         colPromDesc(ifaceSubsystem, "fec_nccw_error_rate", "FEC NCCW Error Rate.", ifacePhysicalLabels),
		"MacsecTxScProtected":                      colPromDesc(ifaceSubsystem, "macsec_output_protected_packets", "Macsec Output Protected.", ifacePhysicalLabels),
		"MacsecTxScEncrypted":                      colPromDesc(ifaceSubsystem, "macsec_output_encrypted_packets", "Macsec Output Encrypted.", ifacePhysicalLabels),
		"MacsecTxScProtectedbytes":                 colPromDesc(ifaceSubsystem, "macsec_output_protectected_bytes", "Macsec Output Protected Bytes.", ifacePhysicalLabels),
		"MacsecTxScEncryptedbytes":                 colPromDesc(ifaceSubsystem, "macsec_output_encrypted_bytes", "Macsec Output Encrypted Bytes.", ifacePhysicalLabels),
		"MacsecRxScOk":                             colPromDesc(ifaceSubsystem, "macsec_input_accepted", "Macsec Input Accepted.", ifacePhysicalLabels),
		"MacsecRxScValidatedbytes":                 colPromDesc(ifaceSubsystem, "macsec_input_validated_bytes", "Macsec Input Validated Bytes.", ifacePhysicalLabels),
		"MacsecRxScDecryptedbytes":                 colPromDesc(ifaceSubsystem, "macsec_input_decrypted_bytes", "Macsec Input Decrypted Bytes.", ifacePhysicalLabels),
		"OversizedFrames":                          colPromDesc(ifaceSubsystem, "multilink_oversized_frames", "Multilink Oversized Frames.", ifacePhysicalLabels),
		"InputErrorFrames":                         colPromDesc(ifaceSubsystem, "multilink_input_error_frames", "Multilink Input Error.", ifacePhysicalLabels),
		"InputDisabledBundle":                      colPromDesc(ifaceSubsystem, "multilink_input_disabled_bundle", "Multilink Input Disabled Bundle.", ifacePhysicalLabels),
		"OutputDisabledBundle":                     colPromDesc(ifaceSubsystem, "multilink_output_disabled_bundle", "Multilink Output Disabled Bundle.", ifacePhysicalLabels),
		"QueuingDrops":                             colPromDesc(ifaceSubsystem, "multilink_queuing_drops", "Multilink Queuing Drops.", ifacePhysicalLabels),
		"PacketBufferOverflow":                     colPromDesc(ifaceSubsystem, "multilink_packet_buffer_overflows", "Multilink Packet Buffer Overflow.", ifacePhysicalLabels),
		"FragmentBufferOverflow":                   colPromDesc(ifaceSubsystem, "multilink_fragment_buffer_overflows", "Multilink Fragment Buffer Overflow.", ifacePhysicalLabels),
		"FragmentTimeout":                          colPromDesc(ifaceSubsystem, "multilink_fragment_timeouts", "Multilink Fragment Timeout.", ifacePhysicalLabels),
		"SequenceNumberMissing":                    colPromDesc(ifaceSubsystem, "multilink_sequence_number_missing", "Multilink Sequence Number Missing.", ifacePhysicalLabels),
		"OutOfOrderSequenceNumber":                 colPromDesc(ifaceSubsystem, "multilink_out_of_order_sequence_number", "Multilink Out of Order Sequence Number.", ifacePhysicalLabels),
		"OutOfRangeSequenceNumber":                 colPromDesc(ifaceSubsystem, "multilink_out_of_range_sequence_number", "Multilink Out of Range Sequence Number.", ifacePhysicalLabels),
		"DataMemoryError":                          colPromDesc(ifaceSubsystem, "multilink_data_memory_errors", "Multilink Data Memory Error.", ifacePhysicalLabels),
		"ControlMemoryError":                       colPromDesc(ifaceSubsystem, "multilink_control_memory_errors", "Multilink Control Memory Error.", ifacePhysicalLabels),
		"FlowErrorAddressSpoofing":                 colPromDesc(ifaceSubsystem, "flow_error_address_spoofing", "Flow Error Address Spoofing.", ifacePhysicalLabels),
		"FlowErrorAuthenticationFailed":            colPromDesc(ifaceSubsystem, "flow_error_authentication_failed", "Flow Error Authentication Failed.", ifacePhysicalLabels),
		"FlowErrorIncomingNat":                     colPromDesc(ifaceSubsystem, "flow_error_incoming_nat", "Flow Error Incoming NAT.", ifacePhysicalLabels),
		"FlowErrorInvalidZone":                     colPromDesc(ifaceSubsystem, "flow_error_invalid_zone", "Flow Error Invalid Zone.", ifacePhysicalLabels),
		"FlowErrorMultipleAuth":                    colPromDesc(ifaceSubsystem, "flow_error_multiple_auth", "Flow Error Multiple Auth.", ifacePhysicalLabels),
		"FlowErrorMultipleIncomingNat":             colPromDesc(ifaceSubsystem, "flow_error_multiple_incoming_nat", "Flow Error Multiple Incoming NAT.", ifacePhysicalLabels),
		"FlowErrorNoGateParent":                    colPromDesc(ifaceSubsystem, "flow_error_no_gate_parent", "Flow Error No Gate Parent.", ifacePhysicalLabels),
		"FlowErrorNoInterestSelfPacket":            colPromDesc(ifaceSubsystem, "flow_error_no_interest_self_packet", "Flow Error No Interest Self Packet.", ifacePhysicalLabels),
		"FlowErrorNoMinorSession":                  colPromDesc(ifaceSubsystem, "flow_error_no_minor_session", "Flow Error No Minor Session.", ifacePhysicalLabels),
		"FlowErrorNoMoreSession":                   colPromDesc(ifaceSubsystem, "flow_error_no_more_session", "Flow Error No More Session.", ifacePhysicalLabels),
		"FlowErrorNoNatGate":                       colPromDesc(ifaceSubsystem, "flow_error_no_nat_gate", "Flow Error No NAT Gate.", ifacePhysicalLabels),
		"FlowErrorNoRoutePresent":                  colPromDesc(ifaceSubsystem, "flow_error_no_route_present", "Flow Error No Route Present.", ifacePhysicalLabels),
		"FlowErrorNoSaForSpi":                      colPromDesc(ifaceSubsystem, "flow_error_no_sa_for_spi", "Flow Error No Security Association for SPI.", ifacePhysicalLabels),
		"FlowErrorNoTunnel":                        colPromDesc(ifaceSubsystem, "flow_error_no_tunnel", "Flow Error No Tunnel.", ifacePhysicalLabels),
		"FlowErrorNoSessionGate":                   colPromDesc(ifaceSubsystem, "flow_error_no_session_gate", "Flow Error No Session Gate.", ifacePhysicalLabels),
		"FlowErrorNullZone":                        colPromDesc(ifaceSubsystem, "flow_error_null_zone", "Flow Error Null Zone.", ifacePhysicalLabels),
		"FlowErrorPolicyDenied":                    colPromDesc(ifaceSubsystem, "flow_error_policy_denied", "Flow Error Policy Denied.", ifacePhysicalLabels),
		"FlowErrorSecurityAssociationMissing":      colPromDesc(ifaceSubsystem, "flow_error_sa_missing", "Flow Error Security Association Missing.", ifacePhysicalLabels),
		"FlowErrorSeqOutsideWindow":                colPromDesc(ifaceSubsystem, "flow_error_seq_outside_window", "Flow Error Seq Outside Window.", ifacePhysicalLabels),
		"FlowErrorSynProtection":                   colPromDesc(ifaceSubsystem, "flow_error_syn_protection", "Flow Error Syn Protection.", ifacePhysicalLabels),
		"FlowErrorUserAuthentication":              colPromDesc(ifaceSubsystem, "flow_error_user_auth", "Flow Error User Auth.", ifacePhysicalLabels),
		"FlowOutputMulticastPackets":               colPromDesc(ifaceSubsystem, "flow_output_multicast_packets", "Flow Output Multicast Packets.", ifacePhysicalLabels),
		"FlowOutputPolicyBytes":                    colPromDesc(ifaceSubsystem, "flow_output_policy_bytes", "Flow Output Policy Bytes.", ifacePhysicalLabels),
		"FlowInputSelfPackets":                     colPromDesc(ifaceSubsystem, "flow_input_self_packets", "Flow Input Self Packets.", ifacePhysicalLabels),
		"FlowInputIcmpPackets":                     colPromDesc(ifaceSubsystem, "flow_input_icmp_packets", "Flow Input ICMP.", ifacePhysicalLabels),
		"FlowInputVpnPackets":                      colPromDesc(ifaceSubsystem, "flow_input_vpn_packets", "Flow Input VPN Packets.", ifacePhysicalLabels),
		"FlowInputMulticastPackets":                colPromDesc(ifaceSubsystem, "flow_input_multicast_packets", "Flow Input Multicast Packets.", ifacePhysicalLabels),
		"FlowInputPolicyBytes":                     colPromDesc(ifaceSubsystem, "flow_input_policy_bytes", "Flow Input Policy Bytes.", ifacePhysicalLabels),
		"FlowInputConnections":                     colPromDesc(ifaceSubsystem, "flow_input_connections", "Flow Input Connections.", ifacePhysicalLabels),
		"SpeedBytes":                               colPromDesc(ifaceSubsystem, "speed_bytes", "Speed of the Interface in Bytes per Second", ifacePhysicalLabels),
		"SnmpIndex":                                colPromDesc(ifaceSubsystem, "snmp_index", "SNMP Index for the interface", ifacePhysicalLabels),
	}
	if len(ifaceDescrKeys) > 0 {
		ifaceDesc["InterfaceDescription"] = colPromDesc(ifaceSubsystem, "description", "Interface description keys", append([]string{"interface"}, ifaceDescrKeys...))
	}
	for _, metricKey := range ifaceMetricKeys {
		ifaceDesc[metricKey] = colPromDesc(ifaceSubsystem, strings.ToLower(metricKey), "User-defined Metric from Description Key", []string{"interface"})
	}
	return ifaceDesc
}

// InterfaceCollector collects Iface metrics, implemented as per the Collector iface.
type InterfaceCollector struct {
	logger log.Logger
}

// NewInterfaceCollector returns a new InterfaceCollector.
func NewInterfaceCollector(logger log.Logger) *InterfaceCollector {
	return &InterfaceCollector{logger: logger}
}

// Name of the collector.
func (*InterfaceCollector) Name() string {
	return ifaceSubsystem
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *InterfaceCollector) Get(ch chan<- prometheus.Metric, conf Config) ([]error, float64) {
	errors := []error{}

	s, err := netconf.DialSSH(conf.SSHTarget, conf.SSHClientConfig)
	if err != nil {
		totalIfaceErrors++
		errors = append(errors, fmt.Errorf("could not connect to %q: %s", conf.SSHTarget, err))
		return errors, totalIfaceErrors
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(`<get-interface-information><extensive/></get-interface-information>`))
	if err != nil {
		totalIfaceErrors++
		errors = append(errors, fmt.Errorf("could not execute netconf RPC call: %s", err))
		return errors, totalIfaceErrors
	}
	if err := processIfaceNetconfReply(reply, ch, conf.IfaceDescrKeys, conf.IfaceMetricKeys, c.logger); err != nil {
		totalIfaceErrors++
		errors = append(errors, err)
	}
	return errors, totalIfaceErrors
}

func (c *BoolIfPresent) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	*c = true
	return nil
}

func processIfaceNetconfReply(reply *netconf.RPCReply, ch chan<- prometheus.Metric, ifaceDescrKeys, ifaceMetricKeys []string, logger log.Logger) error {
	var netconfReply ifaceRPCReply
	r := strings.NewReader(reply.RawReply)
	d := xml.NewDecoder(r)
	if err := d.Decode(&netconfReply); err != nil {
		return fmt.Errorf("could not unmarshal netconf reply xml: %s", err)
	}

	for _, ifaceData := range netconfReply.InterfaceInformation.PhysicalInterface {
		ifaceDesc := getInterfaceDesc(ifaceDescrKeys, ifaceMetricKeys)
		ifaceLabels := []string{strings.TrimSpace(ifaceData.Name.Text)}

		if strings.TrimSpace(ifaceData.AdminStatus.Text) == "up" {
			if strings.TrimSpace(ifaceData.OperStatus.Text) == "up" {
				ch <- prometheus.MustNewConstMetric(ifaceDesc["Up"], prometheus.GaugeValue, 1.0, ifaceLabels...)
			} else {
				ch <- prometheus.MustNewConstMetric(ifaceDesc["Up"], prometheus.GaugeValue, 0.0, ifaceLabels...)
			}
		}
		if len(ifaceData.Speed.Text) > 0 {
			if strings.Contains(strings.TrimSpace(ifaceData.Speed.Text), "Gbps") {
				i, err := strconv.Atoi(strings.TrimRight(strings.TrimSpace(ifaceData.Speed.Text), "Gbps"))
				if err != nil {
					return err
				}
				ch <- prometheus.MustNewConstMetric(ifaceDesc["SpeedBytes"], prometheus.GaugeValue, float64(i*125000000), ifaceLabels...)
			} else if strings.Contains(strings.TrimSpace(ifaceData.Speed.Text), "mbps") {
				i, err := strconv.Atoi(strings.TrimRight(strings.TrimSpace(ifaceData.Speed.Text), "mbps"))
				if err != nil {
					return err
				}
				ch <- prometheus.MustNewConstMetric(ifaceDesc["SpeedBytes"], prometheus.GaugeValue, float64(i*125000), ifaceLabels...)
			}
		}

		var allIfaceDescrKeys map[string]interface{}

		// Junos OS Evolved produces a different representation of a JSON string in the description field.
		// Junos OS Evolved XML output: <description>{\&quot;r_name\&quot;:\&quot;my-far-end-device\&quot;}</description>
		//      XML decoding results in the string: {\\\"r_name\\\":\\\"my-far-end-device\\\"}
		//
		// Junos (regular) XML output: <description>{"r_name":"my-far-end-device"}</description>
		//      XML decoding results in the string: {\"r_name\":\"my-far-end-device\"}
		//
		// This line of code sanitizes the output for the Junos OS Evolved XML response format
		ifaceData.Description.Text = strings.ReplaceAll(ifaceData.Description.Text, "\\\"", "\"")

		if err := json.Unmarshal([]byte(ifaceData.Description.Text), &allIfaceDescrKeys); err != nil {
			allIfaceDescrKeys = nil
		}
		if len(ifaceDescrKeys) > 0 {
			ifaceDescrLabels := []string{strings.TrimSpace(ifaceData.Name.Text)}
			for _, configuredKey := range ifaceDescrKeys {
				if allIfaceDescrKeys[configuredKey] == nil {
					ifaceDescrLabels = append(ifaceDescrLabels, "")
				} else {
					ifaceDescrLabels = append(ifaceDescrLabels, allIfaceDescrKeys[configuredKey].(string))
				}
			}
			newCounter(logger, ch, ifaceDesc["InterfaceDescription"], "1", ifaceDescrLabels...)
		}
		for _, configuredKey := range ifaceMetricKeys {
			if allIfaceDescrKeys[configuredKey] != nil {
				newCounter(logger, ch, ifaceDesc[configuredKey], strings.TrimSpace(allIfaceDescrKeys[configuredKey].(string)), strings.TrimSpace(ifaceData.Name.Text))
			}
		}
		newCounter(logger, ch, ifaceDesc["InterfaceFlapped"], ifaceData.InterfaceFlapped.Seconds, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputBytes"], ifaceData.TrafficStatistics.InputBytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputBytes"], ifaceData.TrafficStatistics.OutputBytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputPackets"], ifaceData.TrafficStatistics.InputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputPackets"], ifaceData.TrafficStatistics.OutputPackets.Text, ifaceLabels...)
		newGauge(logger, ch, ifaceDesc["InputBps"], ifaceData.TrafficStatistics.InputBps.Text, ifaceLabels...)
		newGauge(logger, ch, ifaceDesc["OutputBps"], ifaceData.TrafficStatistics.OutputBps.Text, ifaceLabels...)
		newGauge(logger, ch, ifaceDesc["InputPps"], ifaceData.TrafficStatistics.InputPps.Text, ifaceLabels...)
		newGauge(logger, ch, ifaceDesc["OutputPps"], ifaceData.TrafficStatistics.OutputPps.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["V6InputBytes"], ifaceData.TrafficStatistics.Ipv6TransitStatistics.InputBytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["V6OutputBytes"], ifaceData.TrafficStatistics.Ipv6TransitStatistics.OutputBytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["V6InputPackets"], ifaceData.TrafficStatistics.Ipv6TransitStatistics.InputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["V6OutputPackets"], ifaceData.TrafficStatistics.Ipv6TransitStatistics.OutputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputErrors"], ifaceData.InputErrorList.InputErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputDrops"], ifaceData.InputErrorList.InputDrops.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FramingErrors"], ifaceData.InputErrorList.FramingErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputRunts"], ifaceData.InputErrorList.InputRunts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputGiants"], ifaceData.InputErrorList.InputGiants.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputDiscards"], ifaceData.InputErrorList.InputDiscards.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputResourceErrors"], ifaceData.InputErrorList.InputResourceErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputL3Incompletes"], ifaceData.InputErrorList.InputL3Incompletes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputL2ChannelErrors"], ifaceData.InputErrorList.InputL2ChannelErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputL2MismatchTimeouts"], ifaceData.InputErrorList.InputL2MismatchTimeouts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputFifoErrors"], ifaceData.InputErrorList.InputFifoErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["CarrierTransitions"], ifaceData.OutputErrorList.CarrierTransitions.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputErrors"], ifaceData.OutputErrorList.OutputErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputDrops"], ifaceData.OutputErrorList.OutputDrops.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MtuErrors"], ifaceData.OutputErrorList.MtuErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputResourceErrors"], ifaceData.OutputErrorList.OutputResourceErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputCollisions"], ifaceData.OutputErrorList.OutputCollisions.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["AgedPackets"], ifaceData.OutputErrorList.AgedPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["HsLinkCrcErrors"], ifaceData.OutputErrorList.HsLinkCrcErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputFifoErrors"], ifaceData.OutputErrorList.OutputFifoErrors.Text, ifaceLabels...)
		newGauge(logger, ch, ifaceDesc["SnmpIndex"], ifaceData.SnmpIndex.Text, ifaceLabels...)
		for _, logIface := range ifaceData.LogicalInterfaces {
			logIfaceLabels := []string{strings.TrimSpace(logIface.Name.Text)}
			var allIfaceDescrKeys map[string]interface{}
			if err := json.Unmarshal([]byte(logIface.Description.Text), &allIfaceDescrKeys); err != nil {
				allIfaceDescrKeys = nil
			}
			if logIface.IfConfigFlags.IffUp {
				ch <- prometheus.MustNewConstMetric(ifaceDesc["Up"], prometheus.GaugeValue, 1.0, logIfaceLabels...)
			} else {
				ch <- prometheus.MustNewConstMetric(ifaceDesc["Up"], prometheus.GaugeValue, 0.0, logIfaceLabels...)
			}
			if len(ifaceDescrKeys) > 0 {
				ifaceDescrLabels := []string{strings.TrimSpace(logIface.Name.Text)}

				for _, configuredKey := range ifaceDescrKeys {
					if allIfaceDescrKeys[configuredKey] == nil {
						ifaceDescrLabels = append(ifaceDescrLabels, "")
					} else {
						ifaceDescrLabels = append(ifaceDescrLabels, allIfaceDescrKeys[configuredKey].(string))
					}
				}
				newCounter(logger, ch, ifaceDesc["InterfaceDescription"], "1", ifaceDescrLabels...)
			}
			for _, configuredKey := range ifaceMetricKeys {
				if allIfaceDescrKeys[configuredKey] != nil {
					newCounter(logger, ch, ifaceDesc[configuredKey], strings.TrimSpace(allIfaceDescrKeys[configuredKey].(string)), strings.TrimSpace(logIface.Name.Text))
				}
			}
			trafficStatsSource := logIface.TransitTrafficStatistics
			if logIface.LAGTrafficStatistics.LagBundle.InputBps.Text != "" {
				trafficStatsSource = logIface.LAGTrafficStatistics.LagBundle
			}
			newCounter(logger, ch, ifaceDesc["InputBytes"], trafficStatsSource.InputBytes.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["OutputBytes"], trafficStatsSource.OutputBytes.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["InputPackets"], trafficStatsSource.InputPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["OutputPackets"], trafficStatsSource.OutputPackets.Text, logIfaceLabels...)
			newGauge(logger, ch, ifaceDesc["InputBps"], trafficStatsSource.InputBps.Text, logIfaceLabels...)
			newGauge(logger, ch, ifaceDesc["OutputBps"], trafficStatsSource.OutputBps.Text, logIfaceLabels...)
			newGauge(logger, ch, ifaceDesc["InputPps"], trafficStatsSource.InputPps.Text, logIfaceLabels...)
			newGauge(logger, ch, ifaceDesc["OutputPps"], trafficStatsSource.OutputPps.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["V6InputBytes"], trafficStatsSource.Ipv6TransitStatistics.InputBytes.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["V6OutputBytes"], trafficStatsSource.Ipv6TransitStatistics.OutputBytes.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["V6InputPackets"], trafficStatsSource.Ipv6TransitStatistics.InputPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["V6OutputPackets"], trafficStatsSource.Ipv6TransitStatistics.OutputPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorAddressSpoofing"], logIface.SecurityErrorFlowStatistics.FlowErrorAddressSpoofing.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorAuthenticationFailed"], logIface.SecurityErrorFlowStatistics.FlowErrorAuthenticationFailed.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorIncomingNat"], logIface.SecurityErrorFlowStatistics.FlowErrorIncomingNat.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorInvalidZone"], logIface.SecurityErrorFlowStatistics.FlowErrorInvalidZone.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorMultipleAuth"], logIface.SecurityErrorFlowStatistics.FlowErrorMultipleAuth.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorMultipleIncomingNat"], logIface.SecurityErrorFlowStatistics.FlowErrorMultipleIncomingNat.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoGateParent"], logIface.SecurityErrorFlowStatistics.FlowErrorNoGateParent.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoInterestSelfPacket"], logIface.SecurityErrorFlowStatistics.FlowErrorNoInterestSelfPacket.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoMinorSession"], logIface.SecurityErrorFlowStatistics.FlowErrorNoMinorSession.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoMoreSession"], logIface.SecurityErrorFlowStatistics.FlowErrorNoMoreSession.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoNatGate"], logIface.SecurityErrorFlowStatistics.FlowErrorNoNatGate.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoRoutePresent"], logIface.SecurityErrorFlowStatistics.FlowErrorNoRoutePresent.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoSaForSpi"], logIface.SecurityErrorFlowStatistics.FlowErrorNoSaForSpi.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoTunnel"], logIface.SecurityErrorFlowStatistics.FlowErrorNoTunnel.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNoSessionGate"], logIface.SecurityErrorFlowStatistics.FlowErrorNoSessionGate.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorNullZone"], logIface.SecurityErrorFlowStatistics.FlowErrorNullZone.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorPolicyDenied"], logIface.SecurityErrorFlowStatistics.FlowErrorPolicyDenied.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorSecurityAssociationMissing"], logIface.SecurityErrorFlowStatistics.FlowErrorSecurityAssociationMissing.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorSeqOutsideWindow"], logIface.SecurityErrorFlowStatistics.FlowErrorSeqOutsideWindow.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorSynProtection"], logIface.SecurityErrorFlowStatistics.FlowErrorSynProtection.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowErrorUserAuthentication"], logIface.SecurityErrorFlowStatistics.FlowErrorUserAuthentication.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowInputSelfPackets"], logIface.SecurityInputFlowStatistics.FlowInputSelfPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowInputIcmpPackets"], logIface.SecurityInputFlowStatistics.FlowInputIcmpPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowInputVpnPackets"], logIface.SecurityInputFlowStatistics.FlowInputVpnPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowInputMulticastPackets"], logIface.SecurityInputFlowStatistics.FlowInputMulticastPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowInputPolicyBytes"], logIface.SecurityInputFlowStatistics.FlowInputPolicyBytes.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowInputConnections"], logIface.SecurityInputFlowStatistics.FlowInputConnections.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowOutputMulticastPackets"], logIface.SecurityOutputFlowStatistics.FlowOutputMulticastPackets.Text, logIfaceLabels...)
			newCounter(logger, ch, ifaceDesc["FlowOutputPolicyBytes"], logIface.SecurityOutputFlowStatistics.FlowOutputPolicyBytes.Text, logIfaceLabels...)
			newGauge(logger, ch, ifaceDesc["SnmpIndex"], logIface.SnmpIndex.Text, logIfaceLabels...)
		}
		newCounter(logger, ch, ifaceDesc["StpInputBytesDropped"], ifaceData.StpTrafficStatistics.StpInputBytesDropped.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["StpOutputBytesDropped"], ifaceData.StpTrafficStatistics.StpOutputBytesDropped.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["StpInputPacketsDropped"], ifaceData.StpTrafficStatistics.StpInputPacketsDropped.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["StpOutputPacketsDropped"], ifaceData.StpTrafficStatistics.StpOutputPacketsDropped.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["BitErrorSeconds"], ifaceData.EthernetPcsStatistics.BitErrorSeconds.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["ErroredBlocksSeconds"], ifaceData.EthernetPcsStatistics.ErroredBlocksSeconds.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputBytes"], ifaceData.EthernetMacStatistics.InputBytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputBytes"], ifaceData.EthernetMacStatistics.OutputBytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputPackets"], ifaceData.EthernetMacStatistics.InputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputPackets"], ifaceData.EthernetMacStatistics.OutputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputUnicasts"], ifaceData.EthernetMacStatistics.InputUnicasts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputUnicasts"], ifaceData.EthernetMacStatistics.OutputUnicasts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputBroadcasts"], ifaceData.EthernetMacStatistics.InputBroadcasts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputBroadcasts"], ifaceData.EthernetMacStatistics.OutputBroadcasts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputMulticasts"], ifaceData.EthernetMacStatistics.InputMulticasts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputMulticasts"], ifaceData.EthernetMacStatistics.OutputMulticasts.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputCrcErrors"], ifaceData.EthernetMacStatistics.InputCrcErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputCrcErrors"], ifaceData.EthernetMacStatistics.OutputCrcErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputFifoErrors"], ifaceData.EthernetMacStatistics.InputFifoErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputFifoErrors"], ifaceData.EthernetMacStatistics.OutputFifoErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputMacControlFrames"], ifaceData.EthernetMacStatistics.InputMacControlFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputMacControlFrames"], ifaceData.EthernetMacStatistics.OutputMacControlFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputMacPauseFrames"], ifaceData.EthernetMacStatistics.InputMacPauseFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputMacPauseFrames"], ifaceData.EthernetMacStatistics.OutputMacPauseFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputOversizedFrames"], ifaceData.EthernetMacStatistics.InputOversizedFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputJabberFrames"], ifaceData.EthernetMacStatistics.InputJabberFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputFragmentFrames"], ifaceData.EthernetMacStatistics.InputFragmentFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputVlanTaggedFrames"], ifaceData.EthernetMacStatistics.InputVlanTaggedFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputCodeViolations"], ifaceData.EthernetMacStatistics.InputCodeViolations.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACInputTotalErrors"], ifaceData.EthernetMacStatistics.InputTotalErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MACOutputTotalErrors"], ifaceData.EthernetMacStatistics.OutputTotalErrors.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterInputPackets"], ifaceData.EthernetFilterStatistics.InputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterInputRejectCount"], ifaceData.EthernetFilterStatistics.InputRejectCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterInputRejectDestinationAddressCount"], ifaceData.EthernetFilterStatistics.InputRejectDestinationAddressCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterInputRejectSourceAddressCount"], ifaceData.EthernetFilterStatistics.InputRejectSourceAddressCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterOutputPackets"], ifaceData.EthernetFilterStatistics.OutputPackets.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterOutputPacketPadCount"], ifaceData.EthernetFilterStatistics.OutputPacketPadCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterOutputPacketErrorCount"], ifaceData.EthernetFilterStatistics.OutputPacketErrorCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterCamDestinationFilterCount"], ifaceData.EthernetFilterStatistics.CamDestinationFilterCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FilterCamSourceFilterCount"], ifaceData.EthernetFilterStatistics.CamSourceFilterCount.Text, ifaceLabels...)
		// Some devices can have duplicate traffic class names.
		existingTrafficClasses := make(map[string]int)
		for _, preclStats := range ifaceData.PreclStatistics.PreclInformation.PreclPerClassStatistics {
			trafficClass := strings.TrimSpace(preclStats.PreclTrafficClass.Text)
			if _, exists := existingTrafficClasses[trafficClass]; exists {
				existingTrafficClasses[trafficClass]++
				trafficClass = fmt.Sprintf("%s_%d", trafficClass, existingTrafficClasses[trafficClass])
			} else {
				existingTrafficClasses[trafficClass] = 0
			}
			preclLabels := append(ifaceLabels, trafficClass)
			newCounter(logger, ch, ifaceDesc["PreclRxPackets"], preclStats.PreclRxPackets.Text, preclLabels...)
			newCounter(logger, ch, ifaceDesc["PreclTxPackets"], preclStats.PreclTxPackets.Text, preclLabels...)
			newCounter(logger, ch, ifaceDesc["PreclDroppedPackets"], preclStats.PreclDroppedPackets.Text, preclLabels...)
		}
		newCounter(logger, ch, ifaceDesc["FecCcwCount"], ifaceData.EthernetFecStatistics.FecCcwCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FecNccwCount"], ifaceData.EthernetFecStatistics.FecNccwCount.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FecCcwErrorRate"], ifaceData.EthernetFecStatistics.FecCcwErrorRate.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FecNccwErrorRate"], ifaceData.EthernetFecStatistics.FecNccwErrorRate.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecTxScProtected"], ifaceData.MacsecStatistics.MacsecTxScProtected.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecTxScEncrypted"], ifaceData.MacsecStatistics.MacsecTxScEncrypted.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecTxScProtectedbytes"], ifaceData.MacsecStatistics.MacsecTxScProtectedbytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecTxScEncryptedbytes"], ifaceData.MacsecStatistics.MacsecTxScEncryptedbytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecRxScOk"], ifaceData.MacsecStatistics.MacsecRxScOk.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecRxScValidatedbytes"], ifaceData.MacsecStatistics.MacsecRxScValidatedbytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["MacsecRxScDecryptedbytes"], ifaceData.MacsecStatistics.MacsecRxScDecryptedbytes.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OversizedFrames"], ifaceData.MultilinkInterfaceErrors.OversizedFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputErrorFrames"], ifaceData.MultilinkInterfaceErrors.InputErrorFrames.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["InputDisabledBundle"], ifaceData.MultilinkInterfaceErrors.InputDisabledBundle.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutputDisabledBundle"], ifaceData.MultilinkInterfaceErrors.OutputDisabledBundle.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["QueuingDrops"], ifaceData.MultilinkInterfaceErrors.QueuingDrops.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["PacketBufferOverflow"], ifaceData.MultilinkInterfaceErrors.PacketBufferOverflow.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FragmentBufferOverflow"], ifaceData.MultilinkInterfaceErrors.FragmentBufferOverflow.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["FragmentTimeout"], ifaceData.MultilinkInterfaceErrors.FragmentTimeout.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["SequenceNumberMissing"], ifaceData.MultilinkInterfaceErrors.SequenceNumberMissing.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutOfOrderSequenceNumber"], ifaceData.MultilinkInterfaceErrors.OutOfOrderSequenceNumber.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["OutOfRangeSequenceNumber"], ifaceData.MultilinkInterfaceErrors.OutOfRangeSequenceNumber.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["DataMemoryError"], ifaceData.MultilinkInterfaceErrors.DataMemoryError.Text, ifaceLabels...)
		newCounter(logger, ch, ifaceDesc["ControlMemoryError"], ifaceData.MultilinkInterfaceErrors.ControlMemoryError.Text, ifaceLabels...)
	}
	return nil
}

type ifaceRPCReply struct {
	XMLName              xml.Name         `xml:"rpc-reply"`
	InterfaceInformation ifaceInformation `xml:"interface-information"`
}

type ifaceInformation struct {
	PhysicalInterface []ifacePhysical `xml:"physical-interface"`
}
type ifacePhysical struct {
	Name        ifaceText `xml:"name"`
	AdminStatus ifaceText `xml:"admin-status"`
	OperStatus  ifaceText `xml:"oper-status"`
	Description ifaceText `xml:"description"`
	SnmpIndex   ifaceText `xml:"snmp-index"`
	// Mtu         ifaceText `xml:"mtu"`
	Speed ifaceText `xml:"speed"`
	// LinkType    ifaceText `xml:"link-type"`
	// UpHoldTime               ifaceText                   `xml:"up-hold-time"`
	// DownHoldTime             ifaceText                   `xml:"down-hold-time"`
	// DampHalfLife             ifaceText                   `xml:"damp-half-life"`
	// DampMaxSuppress          ifaceText                   `xml:"damp-max-suppress"`
	// DampReuseLevel           ifaceText                   `xml:"damp-reuse-level"`
	// DampSuppressLevel        ifaceText                   `xml:"damp-suppress-level"`
	// DampSuppressState        ifaceText                   `xml:"damp-suppress-state"`
	// CurrentPhysicalAddress   ifaceText                   `xml:"current-physical-address"`
	// HardwarePhysicalAddress  ifaceText                   `xml:"hardware-physical-address"`
	InterfaceFlapped  ifaceSeconds                `xml:"interface-flapped"`
	TrafficStatistics ifaceInOutBytesPktsBPSPPSV6 `xml:"traffic-statistics"`
	InputErrorList    ifaceInputErrorList         `xml:"input-error-list"`
	OutputErrorList   ifaceOutputErrorList        `xml:"output-error-list"`
	LogicalInterfaces []ifaceLogical              `xml:"logical-interface"`
	// BpduError                ifaceText                   `xml:"bpdu-error"`
	// LdPduError               ifaceText                   `xml:"ld-pdu-error"`
	// L2ptError                ifaceText                   `xml:"l2pt-error"`
	StpTrafficStatistics     ifaceSTPTrafficStats `xml:"stp-traffic-statistics"`
	EthernetPcsStatistics    ifacePCSStats        `xml:"ethernet-pcs-statistics"`
	EthernetMacStatistics    ifaceMACStats        `xml:"ethernet-mac-statistics"`
	EthernetFilterStatistics ifaceFilterStats     `xml:"ethernet-filter-statistics"`
	PreclStatistics          ifacePreclStats      `xml:"precl-statistics"`
	EthernetFecStatistics    ifaceFECStats        `xml:"ethernet-fec-statistics"`
	MacsecStatistics         ifaceMacsecStats     `xml:"macsec-statistics"`
	MultilinkInterfaceErrors ifaceMultilinkErrors `xml:"multilink-interface-errors"`
}

type ifaceMultilinkErrors struct {
	OversizedFrames          ifaceText `xml:"oversized-frames"`
	InputErrorFrames         ifaceText `xml:"input-error-frames"`
	InputDisabledBundle      ifaceText `xml:"input-disabled-bundle"`
	OutputDisabledBundle     ifaceText `xml:"output-disabled-bundle"`
	QueuingDrops             ifaceText `xml:"queuing-drops"`
	PacketBufferOverflow     ifaceText `xml:"packet-buffer-overflow"`
	FragmentBufferOverflow   ifaceText `xml:"fragment-buffer-overflow"`
	FragmentTimeout          ifaceText `xml:"fragment-timeout"`
	SequenceNumberMissing    ifaceText `xml:"sequence-number-missing"`
	OutOfOrderSequenceNumber ifaceText `xml:"out-of-order-sequence-number"`
	OutOfRangeSequenceNumber ifaceText `xml:"out-of-range-sequence-number"`
	DataMemoryError          ifaceText `xml:"data-memory-error"`
	ControlMemoryError       ifaceText `xml:"control-memory-error"`
}

type ifaceMacsecStats struct {
	MacsecTxScProtected      ifaceText `xml:"macsec-tx-sc-protected"`
	MacsecTxScEncrypted      ifaceText `xml:"macsec-tx-sc-encrypted"`
	MacsecTxScProtectedbytes ifaceText `xml:"macsec-tx-sc-protectedbytes"`
	MacsecTxScEncryptedbytes ifaceText `xml:"macsec-tx-sc-encryptedbytes"`
	MacsecRxScOk             ifaceText `xml:"macsec-rx-sc-ok"`
	MacsecRxScValidatedbytes ifaceText `xml:"macsec-rx-sc-validatedbytes"`
	MacsecRxScDecryptedbytes ifaceText `xml:"macsec-rx-sc-decryptedbytes"`
}

type ifaceFECStats struct {
	FecCcwCount      ifaceText `xml:"fec_ccw_count"`
	FecNccwCount     ifaceText `xml:"fec_nccw_count"`
	FecCcwErrorRate  ifaceText `xml:"fec_ccw_error_rate"`
	FecNccwErrorRate ifaceText `xml:"fec_nccw_error_rate"`
}

type ifacePreclStats struct {
	PreclInformation ifacePreclInfo `xml:"precl-information"`
}

type ifacePreclInfo struct {
	PreclPerClassStatistics []ifacePreclPerClassStats `xml:"precl-per-class-statistics"`
}

type ifacePreclPerClassStats struct {
	Text                string    `xml:",chardata"`
	PreclTrafficClass   ifaceText `xml:"precl-traffic-class"`
	PreclRxPackets      ifaceText `xml:"precl-rx-packets"`
	PreclTxPackets      ifaceText `xml:"precl-tx-packets"`
	PreclDroppedPackets ifaceText `xml:"precl-dropped-packets"`
}

type ifacePCSStats struct {
	BitErrorSeconds      ifaceText `xml:"bit-error-seconds"`
	ErroredBlocksSeconds ifaceText `xml:"errored-blocks-seconds"`
}

type ifaceFilterStats struct {
	InputPackets                       ifaceText `xml:"input-packets"`
	InputRejectCount                   ifaceText `xml:"input-reject-count"`
	InputRejectDestinationAddressCount ifaceText `xml:"input-reject-destination-address-count"`
	InputRejectSourceAddressCount      ifaceText `xml:"input-reject-source-address-count"`
	OutputPackets                      ifaceText `xml:"output-packets"`
	OutputPacketPadCount               ifaceText `xml:"output-packet-pad-count"`
	OutputPacketErrorCount             ifaceText `xml:"output-packet-error-count"`
	CamDestinationFilterCount          ifaceText `xml:"cam-destination-filter-count"`
	CamSourceFilterCount               ifaceText `xml:"cam-source-filter-count"`
}

type ifaceMACStats struct {
	InputBytes             ifaceText `xml:"input-bytes"`
	OutputBytes            ifaceText `xml:"output-bytes"`
	InputPackets           ifaceText `xml:"input-packets"`
	OutputPackets          ifaceText `xml:"output-packets"`
	InputUnicasts          ifaceText `xml:"input-unicasts"`
	OutputUnicasts         ifaceText `xml:"output-unicasts"`
	InputBroadcasts        ifaceText `xml:"input-broadcasts"`
	OutputBroadcasts       ifaceText `xml:"output-broadcasts"`
	InputMulticasts        ifaceText `xml:"input-multicasts"`
	OutputMulticasts       ifaceText `xml:"output-multicasts"`
	InputCrcErrors         ifaceText `xml:"input-crc-errors"`
	OutputCrcErrors        ifaceText `xml:"output-crc-errors"`
	InputFifoErrors        ifaceText `xml:"input-fifo-errors"`
	OutputFifoErrors       ifaceText `xml:"output-fifo-errors"`
	InputMacControlFrames  ifaceText `xml:"input-mac-control-frames"`
	OutputMacControlFrames ifaceText `xml:"output-mac-control-frames"`
	InputMacPauseFrames    ifaceText `xml:"input-mac-pause-frames"`
	OutputMacPauseFrames   ifaceText `xml:"output-mac-pause-frames"`
	InputOversizedFrames   ifaceText `xml:"input-oversized-frames"`
	InputJabberFrames      ifaceText `xml:"input-jabber-frames"`
	InputFragmentFrames    ifaceText `xml:"input-fragment-frames"`
	InputVlanTaggedFrames  ifaceText `xml:"input-vlan-tagged-frames"`
	InputCodeViolations    ifaceText `xml:"input-code-violations"`
	InputTotalErrors       ifaceText `xml:"input-total-errors"`
	OutputTotalErrors      ifaceText `xml:"output-total-errors"`
}

type ifaceSTPTrafficStats struct {
	StpInputBytesDropped    ifaceText `xml:"stp-input-bytes-dropped"`
	StpOutputBytesDropped   ifaceText `xml:"stp-output-bytes-dropped"`
	StpInputPacketsDropped  ifaceText `xml:"stp-input-packets-dropped"`
	StpOutputPacketsDropped ifaceText `xml:"stp-output-packets-dropped"`
}

type ifaceLogical struct {
	Name              ifaceText             `xml:"name"`
	Description       ifaceText             `xml:"description"`
	SnmpIndex         ifaceText             `xml:"snmp-index"`
	TrafficStatistics ifaceInOutBytesPktsV6 `xml:"traffic-statistics"`
	IfConfigFlags     ifaceConfigFlags      `xml:"if-config-flags"`
	// LocalTrafficStatistics       ifaceInOutBytesPkts         `xml:"local-traffic-statistics"`
	TransitTrafficStatistics     ifaceInOutBytesPktsBPSPPSV6 `xml:"transit-traffic-statistics"`
	LAGTrafficStatistics         ifaceLAGTrafficStats        `xml:"lag-traffic-statistics"`
	SecurityInputFlowStatistics  ifaceSecInFlow              `xml:"security-input-flow-statistics"`
	SecurityOutputFlowStatistics ifaceSecOutFlow             `xml:"security-output-flow-statistics"`
	SecurityErrorFlowStatistics  ifaceSecErrorFlow           `xml:"security-error-flow-statistics"`
}

type ifaceSecErrorFlow struct {
	FlowErrorAddressSpoofing            ifaceText `xml:"flow-error-address-spoofing"`
	FlowErrorAuthenticationFailed       ifaceText `xml:"flow-error-authentication-failed"`
	FlowErrorIncomingNat                ifaceText `xml:"flow-error-incoming-nat"`
	FlowErrorInvalidZone                ifaceText `xml:"flow-error-invalid-zone"`
	FlowErrorMultipleAuth               ifaceText `xml:"flow-error-multiple-auth"`
	FlowErrorMultipleIncomingNat        ifaceText `xml:"flow-error-multiple-incoming-nat"`
	FlowErrorNoGateParent               ifaceText `xml:"flow-error-no-gate-parent"`
	FlowErrorNoInterestSelfPacket       ifaceText `xml:"flow-error-no-interest-self-packet"`
	FlowErrorNoMinorSession             ifaceText `xml:"flow-error-no-minor-session"`
	FlowErrorNoMoreSession              ifaceText `xml:"flow-error-no-more-session"`
	FlowErrorNoNatGate                  ifaceText `xml:"flow-error-no-nat-gate"`
	FlowErrorNoRoutePresent             ifaceText `xml:"flow-error-no-route-present"`
	FlowErrorNoSaForSpi                 ifaceText `xml:"flow-error-no-sa-for-spi"`
	FlowErrorNoTunnel                   ifaceText `xml:"flow-error-no-tunnel"`
	FlowErrorNoSessionGate              ifaceText `xml:"flow-error-no-session-gate"`
	FlowErrorNullZone                   ifaceText `xml:"flow-error-null-zone"`
	FlowErrorPolicyDenied               ifaceText `xml:"flow-error-policy-denied"`
	FlowErrorSecurityAssociationMissing ifaceText `xml:"flow-error-security-association-missing"`
	FlowErrorSeqOutsideWindow           ifaceText `xml:"flow-error-seq-outside-window"`
	FlowErrorSynProtection              ifaceText `xml:"flow-error-syn-protection"`
	FlowErrorUserAuthentication         ifaceText `xml:"flow-error-user-authentication"`
}

type ifaceSecOutFlow struct {
	FlowOutputMulticastPackets ifaceText `xml:"flow-output-multicast-packets"`
	FlowOutputPolicyBytes      ifaceText `xml:"flow-output-policy-bytes"`
}

type ifaceSecInFlow struct {
	FlowInputSelfPackets      ifaceText `xml:"flow-input-self-packets"`
	FlowInputIcmpPackets      ifaceText `xml:"flow-input-icmp-packets"`
	FlowInputVpnPackets       ifaceText `xml:"flow-input-vpn-packets"`
	FlowInputMulticastPackets ifaceText `xml:"flow-input-multicast-packets"`
	FlowInputPolicyBytes      ifaceText `xml:"flow-input-policy-bytes"`
	FlowInputConnections      ifaceText `xml:"flow-input-connections"`
}

type BoolIfPresent bool
type ifaceConfigFlags struct {
	IffUp BoolIfPresent `xml:"iff-up"`
}

// Leaving the below as it may be implemented in the future
// type ifaceLogicalLocalTrafficStats struct {
// 	InputBytes    ifaceText `xml:"input-bytes"`
// 	OutputBytes   ifaceText `xml:"output-bytes"`
// 	InputPackets  ifaceText `xml:"input-packets"`
// 	OutputPackets ifaceText `xml:"output-packets"`
// }

type ifaceInOutBytesPkts struct {
	InputBytes    ifaceText `xml:"input-bytes"`
	OutputBytes   ifaceText `xml:"output-bytes"`
	InputPackets  ifaceText `xml:"input-packets"`
	OutputPackets ifaceText `xml:"output-packets"`
}

type ifaceInOutBytesPktsV6 struct {
	InputBytes            ifaceText           `xml:"input-bytes"`
	OutputBytes           ifaceText           `xml:"output-bytes"`
	InputPackets          ifaceText           `xml:"input-packets"`
	OutputPackets         ifaceText           `xml:"output-packets"`
	Ipv6TransitStatistics ifaceInOutBytesPkts `xml:"ipv6-transit-statistics"`
}

// Leaving the below as it may be implemented in the future
// type ifaceInOutBytesPktsBPSPPS struct {
// 	InputBytes    ifaceText `xml:"input-bytes"`
// 	OutputBytes   ifaceText `xml:"output-bytes"`
// 	InputPackets  ifaceText `xml:"input-packets"`
// 	OutputPackets ifaceText `xml:"output-packets"`
// 	InputBps      ifaceText `xml:"input-bps"`
// 	OutputBps     ifaceText `xml:"output-bps"`
// 	InputPps      ifaceText `xml:"input-pps"`
// 	OutputPps     ifaceText `xml:"output-pps"`
// }

type ifaceInOutBytesPktsBPSPPSV6 struct {
	InputBytes            ifaceText           `xml:"input-bytes"`
	OutputBytes           ifaceText           `xml:"output-bytes"`
	InputPackets          ifaceText           `xml:"input-packets"`
	OutputPackets         ifaceText           `xml:"output-packets"`
	InputBps              ifaceText           `xml:"input-bps"`
	OutputBps             ifaceText           `xml:"output-bps"`
	InputPps              ifaceText           `xml:"input-pps"`
	OutputPps             ifaceText           `xml:"output-pps"`
	Ipv6TransitStatistics ifaceInOutBytesPkts `xml:"ipv6-transit-statistics"`
}

// Leaving the below as it may be implemented in the future
// type ifaceLogicalTrafficStats struct {
// 	InputBytes            ifaceText           `xml:"input-bytes"`
// 	OutputBytes           ifaceText           `xml:"output-bytes"`
// 	InputPackets          ifaceText           `xml:"input-packets"`
// 	OutputPackets         ifaceText           `xml:"output-packets"`
// 	Ipv6TransitStatistics ifaceInOutBytesPkts `xml:"ipv6-transit-statistics"`
// }

type ifaceLAGTrafficStats struct {
	LagBundle ifaceInOutBytesPktsBPSPPSV6 `xml:"lag-bundle"`
}

type ifaceOutputErrorList struct {
	CarrierTransitions   ifaceText `xml:"carrier-transitions"`
	OutputErrors         ifaceText `xml:"output-errors"`
	OutputDrops          ifaceText `xml:"output-drops"`
	MtuErrors            ifaceText `xml:"mtu-errors"`
	OutputResourceErrors ifaceText `xml:"output-resource-errors"`
	OutputCollisions     ifaceText `xml:"output-collisions"`
	AgedPackets          ifaceText `xml:"aged-packets"`
	HsLinkCrcErrors      ifaceText `xml:"hs-link-crc-errors"`
	OutputFifoErrors     ifaceText `xml:"output-fifo-errors"`
}

type ifaceInputErrorList struct {
	InputErrors             ifaceText `xml:"input-errors"`
	InputDrops              ifaceText `xml:"input-drops"`
	FramingErrors           ifaceText `xml:"framing-errors"`
	InputRunts              ifaceText `xml:"input-runts"`
	InputGiants             ifaceText `xml:"input-giants"`
	InputDiscards           ifaceText `xml:"input-discards"`
	InputResourceErrors     ifaceText `xml:"input-resource-errors"`
	InputL3Incompletes      ifaceText `xml:"input-l3-incompletes"`
	InputL2ChannelErrors    ifaceText `xml:"input-l2-channel-errors"`
	InputL2MismatchTimeouts ifaceText `xml:"input-l2-mismatch-timeouts"`
	InputFifoErrors         ifaceText `xml:"input-fifo-errors"`
}

// Leaving the below as it may be implemented in the future

// type ifaceTrafficStats struct {
// 	InputBytes            ifaceText             `xml:"input-bytes"`
// 	OutputBytes           ifaceText             `xml:"output-bytes"`
// 	InputPackets          ifaceText             `xml:"input-packets"`
// 	OutputPackets         ifaceText             `xml:"output-packets"`
// 	InputBps              ifaceText             `xml:"input-bps"`
// 	OutputBps             ifaceText             `xml:"output-bps"`
// 	InputPps              ifaceText             `xml:"input-pps"`
// 	OutputPps             ifaceText             `xml:"output-pps"`
// 	Ipv6TransitStatistics ifaceIPv6TransitStats `xml:"ipv6-transit-statistics"`
// }

// Leaving the below as it may be implemented in the future
// type ifaceIPv6TransitStats struct {
// 	InputBytes    ifaceText `xml:"input-bytes"`
// 	OutputBytes   ifaceText `xml:"output-bytes"`
// 	InputPackets  ifaceText `xml:"input-packets"`
// 	OutputPackets ifaceText `xml:"output-packets"`
// }

type ifaceText struct {
	Text string `xml:",chardata"`
}

type ifaceSeconds struct {
	Seconds string `xml:"seconds,attr"`
}
