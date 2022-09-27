# check_prometheus

An Icinga check plugin to check Prometheus.

## Usage

### Health

Checks the health or readiness status of the Prometheus server.

* `Health`: Checks the health of an endpoint, which always returns 200 and should be used to check Prometheus health.
* `Ready`: Checks the readiness of an endpoint, which returns 200 when Prometheus is ready to serve traffic (i.e. respond to queries).

````
Usage:
  check_prometheus health

Flags:
  -r, --ready   Checks the readiness of an endpoint
  -i, --info    Displays various build information properties about the Prometheus server
  -h, --help    help for health

Global Flags:
  -H, --hostname string   Address of the prometheus instance (default "localhost")
      --insecure          Allow use of self signed certificates when using SSL
  -p, --port int          Port of the prometheus instance (default 9090)
  -t, --timeout int       Timeout for the check (default 30)
  -S, --tls               Use secure connection
````

````
$ check_prometheus health --hostname 'localhost' --port 9090 --insecure
OK - Prometheus Server is Healthy.

$check_prometheus health --ready       
OK - Prometheus Server is Ready.
````

### Query

```
TODO
```

### Alert

```
check_prometheus alert --name NodeMemoryFull
CRITICAL - Alert NodeMemoryFull firing
exit status 2

check_prometheus alert --name NodeCPULoad
WARNING - Alert NodeCPULoad pending
exit status 1

check_prometheus alert --name NodeDiskFull
OK - Alert NodeDiskFull inactive
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
