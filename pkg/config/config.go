package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/openshift/origin/test/extended/cluster/metrics"
	"github.com/openshift-scale/perf-analyzer/pkg/result"
	"github.com/openshift-scale/perf-analyzer/pkg/utils"
)

type config struct {
	searchDir  string
	resultDir  string
	fileHeader map[string][]string
	hosts      []result.Host
	Metrics    []metrics.Metrics
	keys       []string
}

// NewConfig returns a new configuration struct that contains all fields that we need
func NewConfig(search, result, block, net, process string) config {
	c := config{
		searchDir:  utils.TrailingSlash(search),
		resultDir:  utils.TrailingSlash(result),
		fileHeader: map[string][]string{},
	}
	c.addHeaders(block, net, process)

	return c
}

// addHeaders will check the command line flags to create the files and headers we're looking for
func (c *config) addHeaders(blockString, netString, processString string) {
	blockDevices := strings.Split(blockString, ",")
	// If no block devices were passed, don't add to search
	if len(blockDevices) > 0 {
		c.fileHeader["disk_IOPS.csv"] = blockDevices
	}

	netDevices := strings.Split(netString, ",")
	// If no network devices were passed, don't add to search
	if len(netDevices) > 0 {
		c.fileHeader["network_l2_network_packets_sec.csv"] = netDevices
		c.fileHeader["network_l2_network_Mbits_sec.csv"] = netDevices
	}

	processList := strings.Split(processString, ",")
	// If no process names were passed, don't add to search
	if len(processList) > 0 {
		c.fileHeader["cpu_usage_percent_cpu.csv"] = processList
		c.fileHeader["memory_usage_resident_set_size.csv"] = processList
	}
}

// InitHosts will create the initial host structures with the Kind and ResultDir for each
func (c *config) InitHosts() {
	// This regexp matches the prefix to each pbench host result directory name
	// which indicates host type. (ie. svt-master-1:pbench-benchmark-001/)
	hostRegex := regexp.MustCompile(`svt[_-][ceilmn]\w*[_-]\d`)
	// Return directory listing of searchDir
	dirList, err := ioutil.ReadDir(c.searchDir)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over directory contents
	for _, item := range dirList {
		// Match subdirectory that follows our pattern
		if hostRegex.MatchString(item.Name()) && item.IsDir() {
			Kind := strings.Split(item.Name(), ":")
			newHost := result.Host{
				Kind:      Kind[0],
				ResultDir: c.searchDir + item.Name(),
			}
			c.hosts = append(c.hosts, newHost)
		}
	}
}

// addKeys will help us print a consistent CSV order by sorting our keys
func (c *config) addKeys() {
	// Maps are not ordered, create ordered slice of keys and sort
	// This ensures that the file output is identical between execution
	for k := range c.fileHeader {
		c.keys = append(c.keys, k)
	}
	sort.Strings(c.keys)
}

// Process does the bulk of the math reading the CSV raw data and saving results
func (c *config) Process() {
	c.addKeys()
	for i, host := range c.hosts {
		// Find each raw data CSV
		for _, key := range c.keys {
			fileList := utils.FindFile(host.ResultDir, key)
			// FindFile returns slice, though there should only be one file
			for _, file := range fileList {
				// Parse file into 2d-string slice
				sliceResult, err := utils.ReadCSV(file)
				if err != nil {
					fmt.Printf("Error reading %v: %v\n", file, err)
					continue
				}
				// In a single file we have multiple headers to extract
				for _, header := range c.fileHeader[key] {
					// Extract single column of data that we want
					newResult, err := result.NewSlice(sliceResult, header)
					if err != nil {
						//need to keep list of columns same for all types
						//continue
						fmt.Printf("NewSlice returned error: %v\n", err)
					}

					// Mutate host to add calcuated stats to object
					c.hosts[i].AddResult(newResult, file, header, key)

				}
			}
		}
	}

	var m []metrics.Metrics
	err := utils.GetMetrics(c.searchDir, &m)
	if err != nil {
		fmt.Printf("Error getting Metrics: %v\n", err)
	} else {
		c.Metrics = m
	}
}

// WriteToDisk will write the results to disk as a CSV and a JSON file
func (c *config) WriteToDisk() error {
	err := utils.WriteCSV(c.resultDir, c.keys, c.fileHeader, c.hosts)
	if err != nil {
		return err
	}

	err = utils.WriteJSON(c.resultDir, result.Result{Hosts: c.hosts, Metrics: c.Metrics})
	if err != nil {
		return err
	}
	return nil
}
