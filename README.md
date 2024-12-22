# check_prometheus

An Icinga check plugin to check Prometheus.

## Usage

```bash
Usage:
  check_prometheus [flags]
  check_prometheus [command]

Available Commands:
  alert       Checks the status of a Prometheus alert
  health      Checks the health or readiness status of the Prometheus server
  query       Checks the status of a Prometheus query

Flags:
  -H, --hostname string    Hostname of the Prometheus server (CHECK_PROMETHEUS_HOSTNAME) (default "localhost")
  -p, --port int           Port of the Prometheus server (default 9090)
  -U, --url string         URL/Path to append to the Promethes Hostname (CHECK_PROMETHEUS_URL) (default "/")
  -s, --secure             Use a HTTPS connection
  -i, --insecure           Skip the verification of the server's TLS certificate
  -b, --bearer string      Specify the Bearer Token for server authentication (CHECK_PROMETHEUS_BEARER)
  -u, --user string        Specify the user name and password for server authentication <user:password> (CHECK_PROMETHEUS_BASICAUTH)
      --ca-file string     Specify the CA File for TLS authentication (CHECK_PROMETHEUS_CA_FILE)
      --cert-file string   Specify the Certificate File for TLS authentication (CHECK_PROMETHEUS_CERT_FILE)
      --key-file string    Specify the Key File for TLS authentication (CHECK_PROMETHEUS_KEY_FILE)
  -t, --timeout int        Timeout in seconds for the CheckPlugin (default 30)
  -h, --help               help for check_prometheus
  -v, --version            version for check_prometheus
```

The check plugin respects the environment variables `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY`.

Various flags can be set with environment variables, refer to the help to see which flags.

In the case Prometheus runs behind a reverse proxy, the `--url` parameter can be used:

```bash
# https://monitoring.example.com:443/subpath

$ check_prometheus health -H 'monitoring.example.com' --port 443 --secure --url /subpath
OK - Prometheus Server is Healthy. | statuscode=200
```

### Health

Checks the health or readiness status of the Prometheus server.

* `Health`: Checks the health of an endpoint, which returns OK if the Prometheus server is healthy.
* `Ready`: Checks the readiness of an endpoint, which returns OK if the Prometheus server is ready to serve traffic (i.e. respond to queries).

```bash
Usage:
  check_prometheus health [flags]

Examples:
  $ check_prometheus health --hostname 'localhost' --port 9090 --insecure
  OK - Prometheus Server is Healthy. | statuscode=200

Flags:
  -r, --ready   Checks the readiness of an endpoint
  -I, --info    Displays various build information properties about the Prometheus server
  -h, --help    help for health
```

```bash
$ check_prometheus health --hostname 'localhost' --port 9090 --insecure
OK - Prometheus Server is Healthy. | statuscode=200

$ check_prometheus health --ready
OK - Prometheus Server is Ready. | statuscode=200
```

### Query

Checks the status of a Prometheus query and evaluates the result of the alert.

>Note: Time range values e.G. 'go_memstats_alloc_bytes_total[10s]', only the latest value will be evaluated, other values will be ignored!

```bash
Usage:
  check_prometheus query [flags]

Examples:
  $ check_prometheus query -q 'go_gc_duration_seconds_count' -c 5000 -w 2000
  CRITICAL - 2 Metrics: 1 Critical - 0 Warning - 1 Ok
   \_[OK] go_gc_duration_seconds_count{instance="localhost:9090", job="prometheus"} - value: 1599
   \_[CRITICAL] go_gc_duration_seconds_count{instance="node-exporter:9100", job="node-exporter"} - value: 79610
   | value_go_gc_duration_seconds_count_localhost:9090_prometheus=1599 value_go_gc_duration_seconds_count_node-exporter:9100_node-exporter=79610

Flags:
  -q, --query string      An Prometheus query which will be performed and the value result will be evaluated
  -w, --warning string    The warning threshold for a value (default "10")
  -c, --critical string   The critical threshold for a value (default "20")
  -h, --help              help for query
```

#### Checking a single metric with ONE direct vector result

```bash
$ check_prometheus query -q 'go_goroutines{job="prometheus"}' -c 40 -w 27
WARNING - 1 Metrics: 0 Critical - 1 Warning - 0 Ok
 \_[WARNING] go_goroutines{instance="localhost:9090", job="prometheus"} - value: 37
 | value_go_goroutines_localhost:9090_prometheus=37
```

#### Checking a single metric with multiple vector results

```bash
$ check_prometheus query -q 'go_goroutines' -c 40 -w 27
WARNING - 2 Metrics: 0 Critical - 1 Warning - 1 Ok
 \_[WARNING] go_goroutines{instance="localhost:9090", job="prometheus"} - value: 37
 \_[OK] go_goroutines{instance="node-exporter:9100", job="node-exporter"} - value: 7
 | value_go_goroutines_localhost:9090_prometheus=37 value_go_goroutines_node-exporter:9100_node-exporter=7
```

#### Checking a time series matrix result

Hint: Currently only the latest value will be evaluated, other values will be ignored.

```bash
$ check_prometheus query -q 'go_goroutines{job="prometheus"}[10s]' -c5 -w 10
CRITICAL - 1 Metrics: 1 Critical - 0 Warning - 0 Ok
 \_[CRITICAL] go_goroutines{instance="localhost:9090", job="prometheus"} - value: 37
 | value_go_goroutines_localhost:9090_prometheus=37

$ check_prometheus query -q 'go_goroutines[10s]' -c 50 -w 40
OK - 2 Metrics OK | value_go_goroutines_localhost:9090_prometheus=37 value_go_goroutines_node-exporter:9100_node-exporter=7
```

### Alert

Checks the status of a Prometheus alert and evaluates the status of the alert.

```bash
Usage:
  check_prometheus alert [flags]

Examples:
  $ check_prometheus alert --name "PrometheusAlertmanagerJobMissing"
  CRITICAL - 1 Alerts: 1 Firing - 0 Pending - 0 Inactive
   \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] is firing - value: 1.00
   | firing=1 pending=0 inactive=0

  $ check_prometheus a alert --name "PrometheusAlertmanagerJobMissing" --name "PrometheusTargetMissing"
  CRITICAL - 2 Alerts: 1 Firing - 0 Pending - 1 Inactive
   \_[OK] [PrometheusTargetMissing] is inactive
   \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] is firing - value: 1.00
   | total=2 firing=1 pending=0 inactive=1

Flags:
      --exclude-alert stringArray  Alerts to ignore. Can be used multiple times and supports regex.
  -h, --help                       help for alert
  -n, --name strings               The name of one or more specific alerts to check.
                                   This parameter can be repeated e.G.: '--name alert1 --name alert2'
                                   If no name is given, all alerts will be evaluated
  -T, --no-alerts-state string     State to assign when no alerts are found (0, 1, 2, 3, OK, WARNING, CRITICAL, UNKNOWN). If not set this defaults to OK (default "OK")
  -P, --problems                   Display only alerts which status is not inactive/OK. Note that in combination with the --name flag this might result in no alerts being displayed
```

#### Checking all defined alerts

```bash
$ check_prometheus alert
CRITICAL - 6 Alerts: 3 Firing - 0 Pending - 3 Inactive
 \_[OK] [PrometheusTargetMissing] is inactive
 \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] is firing - value: 1.00
 \_[OK] [HostOutOfMemory] - Job: [alertmanager]
 \_[OK] [HostHighCpuLoad] - Job: [alertmanager]
 \_[CRITICAL] [HighResultLatency] - Job: [prometheus] on Instance: [localhost:9090]  is firing - value: 11.00
 \_[CRITICAL] [HighResultLatency] - Job: [node-exporter] on Instance: [node-exporter:9100]  is firing - value: 10.00
 | total=6 firing=3 pending=0 inactive=3

```

#### Checking multiple alerts

```bash
$ check_prometheus alert --name "HostHighCpuLoad" --name "HighResultLatency"
CRITICAL - 3 Alerts: 2 Firing - 0 Pending - 1 Inactive
 \_[OK] [HostHighCpuLoad] is inactive
 \_[CRITICAL] [HighResultLatency] - Job: [prometheus] on Instance: [localhost:9090]  is firing - value: 11.00
 \_[CRITICAL] [HighResultLatency] - Job: [node-exporter] on Instance: [node-exporter:9100]  is firing - value: 10.00
 | total=3 firing=2 pending=0 inactive=1
```

```bash
$ check_prometheus alert --name "HostHighCpuLoad" --name "PrometheusTargetMissing"
OK - Alerts inactive | total=2 firing=0 pending=0 inactive=2
```

### Special cases

#### Your Prometheus runs behind a reverse proxy

>Example: <https://monitoring.example.com:443/subpath>

```bash
$ check_prometheus health --hostname 'monitoring.example.com' --port 443 --secure --url /subpath
OK - Prometheus Server is Healthy. | statuscode=200
```

## License

Copyright (c) 2022 [NETWAYS GmbH](mailto:info@netways.de)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not,
see [gnu.org/licenses](https://www.gnu.org/licenses/).
