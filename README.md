# perf-analyzer

CI-friendly golang tool for parsing the output of time series data created by [pbench](https://github.com/distributed-system-analysis/pbench) in CSV format, as well as from [prometheus](https://github.com/prometheus/prometheus) directly.

```
Usage of ./scraper:
  -blkdev string
        List of block devices (default "sda-write,sda-read,vda-write,vda-read,xvda-write,xvda-read,xvdb-write,xvdb-read,nvme0n1-write,nvme0n1-read")
  -duration int
        Duration of test in integer minutes (used to calculate quest start time) (default 30)
  -i string
        pbench run result directory to parse (default "/var/lib/pbench-agent/benchmark_result/tools-default/")
  -insecure
        Trust self-signed HTTP certificates
  -netdev string
        List of network devices (default "eth0-rx,eth0-tx")
  -o string
        output directory for parsed CSV result data (default "/tmp/")
  -pbench
        scrape pbench results
  -proc string
        list of processes to gather (default "openshift_start_master_api_,openshift_start_master_controll,hyperkube_kubelet_,openshift_start_node_,etcd,dockerd-current_,elasticsearc,prometheus_,systemd_--switched-root,openshift_start_network_,ovs-vswitchd_unix,openshift-router,fluentd,kibana,heapster,crio")
  -prometheus
        scrape prometheus endpoint
  -step string
        Query resolution step width in number of seconds (default "1m")
  -token string
        Authorization type + token for endpoint
  -url string
        URL for prometheus connection (default "http://localhost:9090")
```

Example pbench command:
```
./scraper -pbench -i ~/work/pbench-result/tools-default/ -o ~/data/ -blkdev vda-write -blkdev xvdb -netdev eth0-rx -netdev eth0-tx
```

`blkdev` represents a single block device name, to add more than one block device, you will need to pass the flag again per device, as above

`i` is the input directory, it must point to the parent of the host data, which is `.../tools-default/`

`o` is the output directory, it can be any directory (dirname/)

`netdev` represents a single network device name, to add more than more network device, you will need to pass the flag again per device, as above

`proc` is a comma-separated list of process names to extract results for, avoid spaces

## Prometheus Usage

If you intend to scrape prometheus you must use the `-prometheus` flag to enable. Prometheus queries have one mandatory flag: `-url`.

To test the tool against an OpenShift cluster try this test script:

```
#!/bin/bash
#
# Test prometheus scraping
#

PROMETHEUS_URL=https://$(oc --config config get route -n openshift-monitoring | grep prometheus | awk '{print $2}')
AUTH_TOKEN=$(oc --config config sa get-token prometheus-k8s -n openshift-monitoring)

./_output/scraper -prometheus -url ${PROMETHEUS_URL} -insecure -token "${AUTH_TOKEN}"
```

Since we're using OpenShift we need to add the bearer token (`-token`) for the proxy authentication. Also the certs are self signed so we need to disable TLS verification.
