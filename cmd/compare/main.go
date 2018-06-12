package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/redhat-performance/pbench-analyzer/pkg/result"
)

var oldFile, newFile string
var stdDev float64

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
		fmt.Fprintf(os.Stderr, "Error reading file \"%v\"\n", oldFile)
		return
	}

	newRun, err := readResultJSON(newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file \"%v\"\n", newFile)
		return
	}

	if !areHostResultsSimilar(oldRun.Hosts, newRun.Hosts) {
		fmt.Fprintf(os.Stderr, "Results do not have the same number of hosts and/or results per host.\n")
		return
	}

	for i := range oldRun.Hosts {
		for j := range oldRun.Hosts[i].Results {
			if newRun.Hosts[i].Results[j].Pct95 > oldRun.Hosts[i].Results[j].Pct95*(1+stdDev) ||
				newRun.Hosts[i].Results[j].Pct95 < oldRun.Hosts[i].Results[j].Pct95*(1-stdDev) {
				fmt.Printf("%s: Out of spec %s process with %s\n", newRun.Hosts[i].Kind, newRun.Hosts[i].Results[j].Kind, newRun.Hosts[i].Results[j].Resource)
				return
			}
		}
	}

	// TODO compare metrics?
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
