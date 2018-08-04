# Junos Exporter
A Prometheus exporter that collects metrics from Junos devices using SSH NETCONF sessions and exposes them via HTTP, ready for collection by Prometheus.

## Getting Started
Start junos_exporter with a valid configuration file using the --config.path flag. To then collect the metrics of a device, pass the 'config' and 'target' parameter to the exporter's web interface. For example, http://exporter:9347/metrics?config=default&target=192.168.1.1.

Promethues configuraiton:
```
scrape_configs:
  - job_name: junos
    static_configs:
      - targets:
        - device1
    params:
      config: ['default']
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: junos_exporter:9347  # Junos exporter's address and port.
```

Docker:
```
docker run --restart unless-stopped -d -p 9347:9347 -v /home/user/.ssh/ssh_key:/ssh_key  -v /home/user/config.yaml:/config.yaml tynany/junos_exporter
```
The above Docker commands assumes a configuration file that specifies the SSK key as /ssh_key is located locally in /home/user/config.yaml.

## Configuration file
Junos Exporter requires a configuration file in the below format:
```
configs:
  default:                  # Name of the configuration
    timeout:                # SSH Timeout in seconds. Optional.
    username:               # SSH Username. Required.
    password:               # SSH Password. Optional.    
    ssh_key:                # SSH Key. Optional.        
    allowed_targets:        # List of targets that can be collected. Optional.
      -          
    enabled_collectors:     # Which collectors to enable. Required.
      - bgp
      - interface
      - environment
      - power
global:
  timeout:                  # SSH Timeout in seconds, globally configured. Optional.
  allowed_targets:          # List of targets that can be collected, globally configured. Optional.
   -
```
### Example
```
configs:
  default:
    timeout: 30
    username: user1
    password: securepassword
    ssh_key: ~/key.pem
    allowed_targets:
      - 10.1.0.0
    enabled_collectors:
      - bgp
      - interface
  bgp_only:
    username: user2
    password: password
    enabled_collectors:
      - bgp
global:
  timeout: 30
  allowed_targets:
   - 10.0.0.0

```
### configs
Each configuration is called by passing the 'config' parameter to the exporter's export web interface. In the above example, to use the default config you would use http://exporter:9347/metrics?config=default and to use the bgp_only config, you would use http://exporter:9347/metrics?config=bgp_only.

### allowed_targets
If allowed_targets is specified, only those targets may be collected. This is a form of security that stops a malicious user trying to collect details, such as the username and password, by specifying a target they control.

### global
Global applies to all configs, where that configuration item has not already been set under a specific config.

## Metrics
The below metrics are currently implemented.
- Interface Statistics, from `show interface extensive`.
- BGP, from `show bgp summary`.
- Environment, from `show chassis environment`.
- Power, from `show chassis power detail`.

### BGP: junos_bgp_peer_types_up
Junos Exporter exposes a special metric, `junos_bgp_peer_types_up`, that can be used in scenarios where you want to create Prometheus queries that report on the number of types of BGP peers that are currently established, such as for Alert Manager. To implement this metric, a JSON formatted description with a 'type' element must be configured on your BGP group. Junos Exporter will then aggregate all BGP peers that are currently established and configured with that type.

For example, if you want to know how many BGP peers are currently established that provide internet, you'd set the description of all BGP groups that provide internet to `{"type":"internet"}` and query Prometheus with `junos_bgp_peer_types_up{type="internet"})`. Going further, if you want to create an alert when the number of established BGP peers that provide internet is 1 or less, you'd use `sum(junos_bgp_peer_types_up{type="internet"}) <= 1`.  

Example Junos configuration:
```
set protocols bgp group internet-provider1 description "{\"type\":\"internet\"}"
set protocols bgp group internet-provider2 description "{\"type\":\"internet\"}"
```

## Development
### Building
```
go get github.com/tynany/junos_exporter
cd ${GOPATH}/src/github.com/tynany/junos_exporter
go build
```

### NETCONF Output
XML was chosen as the output format of NETCONF commands for the below reasons:

 - Junos devices return XML faster than any other format, more than half the time it takes fora JSON response. Presumedly this is because XML is the native configuration format of Junos.
 - It is only possible to use NETCONF filter tags when the output is XML.
