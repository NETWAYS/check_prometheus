# check_prometheus

An Icinga check plugin to check Prometheus.

## Usage

### Health

Checks the health or readiness status of the Prometheus server.

* `Health`: Checks the health of an endpoint, which always returns 200 and should be used to check Prometheus health.
* `Ready`: Checks the readiness of an endpoint, which returns 200 when Prometheus is ready to serve traffic (i.e. respond to queries).

````
Usage:
  check_prometheus [flags]
  check_prometheus [command]

Available Commands:
  alert       Checks the status of an Prometheus alert
  health      Checks the health or readiness status of the Prometheus server
  query       Checks the status of an Prometheus query

Flags:
  -H, --hostname string    Hostname of the Prometheus server (default "localhost")
  -p, --port int           Port of the Prometheus server (default 9090)
  -s, --secure             Use a HTTPS connection
  -i, --insecure           Skip the verification of the server's TLS certificate
  -b, --bearer string      Specify the Bearer Token for server authentication
  -u, --user string        Specify the user name and password for server authentication <user:password>
      --ca-file string     Specify the CA File for TLS authentication
      --cert-file string   Specify the Certificate File for TLS authentication
      --key-file string    Specify the Key File for TLS authentication
  -t, --timeout int        Timeout in seconds for the CheckPlugin (default 30)
  -h, --help               help for check_prometheus
  -v, --version            version for check_prometheus
````

````
$ check_prometheus health --hostname 'localhost' --port 9090 --insecure
OK - Prometheus Server is Healthy.

$check_prometheus health --ready
OK - Prometheus Server is Ready.
````

### Query

#### Checking a single metric with ONE direct vector result
```
$ check_prometheus query -n 'go_goroutines{job="prometheus"}' -c 40 -w 27
WARNING - go_goroutines{instance="localhost:9090", job="prometheus"}
 \_ 37 @ 2022-10-12 14:03:53.256 +0200 CEST
```

#### Checking a single metric with multiple vector results
````
$ check_prometheus query -n 'go_goroutines' -c 40 -w 27 
WARNING - Found 2 Metrics - 0 Critical - 1 Warning - 1 Ok
[WARNING] go_goroutines{instance="localhost:9090", job="prometheus"}
 \_ 37 @ 2022-10-12 14:05:34.744 +0200 CEST
[OK] go_goroutines{instance="node-exporter:9100", job="node-exporter"}
 \_ 37 @ 2022-10-12 14:05:34.744 +0200 CEST
 \_ 7 @ 2022-10-12 14:05:34.744 +0200 CEST
````
#### Checking a timeseries matrix result
````
$ check_prometheus query -n 'go_goroutines{job="prometheus"}[10s]' -c5 -w 10
CRITICAL - [CRITICAL] go_goroutines{instance="localhost:9090", job="prometheus"}
 \_ 37 @ 2022-10-12 14:15:27.45 +0200 CEST
 \_ 37 @ 2022-10-12 14:15:32.451 +0200 CEST

$ check_prometheus query -n 'go_goroutines{job="prometheus"}[10s]' -c 50 -w 40
OK - Found 1 Metrics - all Metrics Ok
````

### Alert

```
$ check_prometheus alert
CRITICAL - Found 6 alerts - firing 3 - pending 0 - inactive 3
 \_[OK] [PrometheusTargetMissing] is inactive 
 \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] on Instance: [] is firing 
 \_[OK] [HostOutOfMemory] - Job: [alertmanager] on Instance: [] is inactive 
 \_[OK] [HostHighCpuLoad] - Job: [alertmanager] on Instance: [] is inactive 
 \_[CRITICAL] [HighResultLatency] - Job: [prometheus] on Instance: [localhost:9090] is firing 
 \_[CRITICAL] [HighResultLatency] - Job: [node-exporter] on Instance: [node-exporter:9100] is firing 
 
$ check_prometheus alert --name "HostHighCpuLoad" --name "HighResultLatency" 
CRITICAL - Found 3 alerts - firing 2 - pending 0 - inactive 1
 \_[OK] [HostHighCpuLoad] is inactive 
 \_[CRITICAL] [HighResultLatency] - Job: [prometheus] on Instance: [localhost:9090] is firing 
 \_[CRITICAL] [HighResultLatency] - Job: [node-exporter] on Instance: [node-exporter:9100] is firing 
 
$ check_prometheus alert --name "HostHighCpuLoad" --name "PrometheusTargetMissing"
OK - All alerts are inactive
```

## License

Copyright (c) 2022 [NETWAYS GmbH](mailto:info@netways.de)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not,
see [gnu.org/licenses](https://www.gnu.org/licenses/).
