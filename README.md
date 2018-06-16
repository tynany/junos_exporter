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
      config: default
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: junos_exporter:9347  # Junos exporter's address and port.
```

## Configuration file
Junos exporter requires a configuration file in the below format:
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

## Development
### Building
```
go get github.com/tynany/junos_exporter
cd ${GOPATH}/src/github.com/tynany/junos_exporter
go build
```
