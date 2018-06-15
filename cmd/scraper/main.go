package main

import (
	"flag"
	"fmt"

	"github.com/redhat-performance/pbench-analyzer/pkg/config"
)

var searchDir, resultDir, processString, netString, blockString string

func initFlags() {
	flag.StringVar(&searchDir, "i", "/var/lib/pbench-agent/benchmark_result/tools-default/", "pbench run result directory to parse")
	flag.StringVar(&resultDir, "o", "/tmp/", "output directory for parsed CSV result data")
	flag.StringVar(&processString, "proc", "openshift_start_master_api_,openshift_start_master_controllers_,hyperkube_kubelet_,etcd,dockerd-current_,elasticsearc,prometheus_,systemd_--switched-root,openshift_start_network_,ovs-vswitchd_unix,openshift-router,fluentd,kibana,heapster,crio", "list of processes to gather")
	flag.StringVar(&blockString, "blkdev", "sda-write,sda-read,vda-write,vda-read,xvda-write,xvda-read,xvdb-write,xvdb-read,nvme0n1-write,nvme0n1-read", "List of block devices")
	flag.StringVar(&netString, "netdev", "eth0-rx,eth0-tx", "List of network devices")
	flag.Parse()
}

func main() {
	initFlags()

	// Check if no flags were passed, print help
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		return
	}

	// Create new config structure which will contain all data
	c := config.NewConfig(searchDir, resultDir, blockString, netString, processString)

	// Initialize each host struct
	c.InitHosts()

	// Process results for each host
	c.Process()

	// Write CSV and JSON to disk
	err := c.WriteToDisk()
	if err != nil {
		fmt.Printf("Error writing files to disk: %v", err)
	}
}
