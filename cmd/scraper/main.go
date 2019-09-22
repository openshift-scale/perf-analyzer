package main

import (
	"flag"
	"fmt"

	"github.com/openshift-scale/perf-analyzer/pkg/config"
	"github.com/openshift-scale/perf-analyzer/pkg/prometheus"
)

func initFlags() (cfg config.ScrapeConfig) {
	flag.BoolVar(&cfg.EnablePbenchFlag, "pbench", false, "scrape pbench results")
	flag.BoolVar(&cfg.EnablePrometheusFlag, "prometheus", false, "scrape prometheus endpoint")
	flag.BoolVar(&cfg.InsecureTLSFlag, "insecure", false, "Trust self-signed HTTP certificates")
	flag.IntVar(&cfg.DurationFlag, "duration", 30, "Duration of test in integer minutes (used to calculate quest start time)")
	flag.StringVar(&cfg.StepFlag, "step", "1m", "Query resolution step width in number of seconds")
	flag.StringVar(&cfg.TokenFlag, "token", "", "Authorization type + token for endpoint")
	flag.StringVar(&cfg.UrlFlag, "url", "http://localhost:9090", "URL for prometheus connection")
	flag.StringVar(&cfg.SearchDir, "i", "/var/lib/pbench-agent/benchmark_result/tools-default/", "pbench run result directory to parse")
	flag.StringVar(&cfg.ResultDir, "o", "/tmp/", "output directory for parsed CSV result data")
	flag.StringVar(&cfg.ProcessString, "proc", "openshift_start_master_api_,openshift_start_master_controll,hyperkube_kubelet_,openshift_start_node_,etcd,dockerd-current_,elasticsearc,prometheus_,systemd_--switched-root,openshift_start_network_,ovs-vswitchd_unix,openshift-router,fluentd,kibana,heapster,crio", "list of processes to gather")
	flag.StringVar(&cfg.BlockString, "blkdev", "sda-write,sda-read,vda-write,vda-read,xvda-write,xvda-read,xvdb-write,xvdb-read,nvme0n1-write,nvme0n1-read", "List of block devices")
	flag.StringVar(&cfg.NetString, "netdev", "eth0-rx,eth0-tx", "List of network devices")
	flag.Parse()

	return
}

func main() {
	cfg := initFlags()

	// Check if no flags were passed, print help
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		return
	}

	if cfg.EnablePrometheusFlag {
		prometheus.DoPrometheusQuery(cfg)
	}

	if cfg.EnablePbenchFlag {
		// Create new config structure which will contain all data
		c := config.NewConfig(cfg)

		// Initialize each host struct
		c.Init()

		// Process results for each host
		c.Process()

		// Write CSV and JSON to disk
		err := c.WriteToDisk()
		if err != nil {
			fmt.Printf("Error writing files to disk: %v", err)
		}
	}
}
