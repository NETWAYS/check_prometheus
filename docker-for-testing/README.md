### Simple setup
````
$ docker-compose up -d
````

`altertmanager`: http://localhost:9093/#/alerts

`prometheus`: http://localhost:9090/graph

`node-exporter`: http://localhost:9100/metrics

#### Exmaples

````
$ curl -X PUT  localhost:9090/-/reload
````
````
$ curl localhost:9090/-/healthy
Prometheus Server is Healthy.
````

````
$ curl localhost:9090/-/ready
Prometheus Server is Ready.
````

````
curl -g 'http://localhost:9090/api/v1/series?' --data-urlencode 'match[]=up' --data-urlencode 'match[]=process_start_time_seconds{job="prometheus"}'

{"status":"success","data":[{"__name__":"process_start_time_seconds","instance":"localhost:9090","job":"prometheus"},{"__name__":"up","instance":"localhost:9090","job":"prometheus"},{"__name__":"up","instance":"node-exporter:9100","job":"node-exporter"}]}
(base)
````