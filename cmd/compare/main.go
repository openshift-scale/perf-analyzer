package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openshift-scale/perf-analyzer/pkg/result"
)

var oldFile, newFile string
var stdDev float64
var procAlias map[string]string

func initFlags() {
	flag.StringVar(&oldFile, "old", "", "Previous run summary")
	flag.StringVar(&newFile, "new", "", "New run summary")
	flag.Float64Var(&stdDev, "stddev", 0.05, "Float percentage standard deviation for result tolerance (0.05 = 5%)")
	flag.Parse()
}

func main() {
	initFlags()
	var files []string
	files = append(files, oldFile)
	files = append(files, newFile)
	procAlias = make(map[string]string)
	procAlias["openshift_start_node_"] = "hyperkube_kubelet_"

	if oldFile == "" && newFile == "" {
		fmt.Fprintf(os.Stderr, "Must specify both old and new run data:\n")
		flag.PrintDefaults()
		return
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "File does not exist: %s\n", file)
			return
		}
	}

	oldRun, err := readResultJSON(oldFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file \"%v\": %s\n", oldFile, err)
		return
	}

	newRun, err := readResultJSON(newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file \"%v\": %s\n", newFile, err)
		return
	}

	for i := range oldRun.Hosts {
		k, err := getHostIndex(newRun.Hosts, oldRun.Hosts[i].Kind)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}
		for j := range oldRun.Hosts[i].Results {
			l, err := getResultIndex(newRun.Hosts[k], oldRun.Hosts[i].Results[j])
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				continue
			}
			if newRun.Hosts[k].Results[l].Pct95 > oldRun.Hosts[i].Results[j].Pct95*(1+stdDev) ||
				newRun.Hosts[k].Results[l].Pct95 < oldRun.Hosts[i].Results[j].Pct95*(1-stdDev) {
				fmt.Printf("%s: Out of spec %s process with %s, old: %.2f => new: %.2f\n", newRun.Hosts[k].Kind, newRun.Hosts[k].Results[l].Kind, newRun.Hosts[k].Results[l].Resource, oldRun.Hosts[i].Results[j].Pct95, newRun.Hosts[k].Results[l].Pct95)
			}
		}
	}

	// TODO compare metrics?
}

func getHostIndex(hostResult []result.Host, kind string) (int, error) {
	for h := range hostResult {
		if hostResult[h].Kind == kind {
			return h, nil
		}
	}
	return 0, fmt.Errorf("Host type %s not found.", kind)
}

func getResultIndex(hostResult result.Host, resultItem result.ResultType) (int, error) {
	for r := range hostResult.Results {
		if (hostResult.Results[r].Kind == resultItem.Kind ||
			hostResult.Results[r].Kind == procAlias[resultItem.Kind]) &&
			hostResult.Results[r].Resource == resultItem.Resource {
			return r, nil
		}
	}
	return 0, fmt.Errorf("Result index for %s, %s not found", resultItem.Kind, resultItem.Resource)

}

func readResultJSON(file string) (*result.Result, error) {
	var res result.Result
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func areHostResultsSimilar(old, new []result.Host) bool {
	if len(old) != len(new) {
		return false
	}
	for i := range old {
		if len(old[i].Results) != len(new[i].Results) ||
			old[i].Kind != new[i].Kind {
			return false
		}
		for j := range old[i].Results {
			if old[i].Results[j].Kind != new[i].Results[j].Kind ||
				old[i].Results[j].Resource != new[i].Results[j].Resource {
				return false
			}
		}
	}
	return true
}
