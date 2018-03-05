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

	if !areResultsSimilar(oldRun, newRun) {
		fmt.Fprintf(os.Stderr, "Results do not have the same number of hosts and/or results per host.\n")
		return
	}

	for i := range oldRun {
		for j := range oldRun[i].Results {
			if newRun[i].Results[j].Pct95 > oldRun[i].Results[j].Pct95*(1+stdDev) ||
				newRun[i].Results[j].Pct95 < oldRun[i].Results[j].Pct95*(1-stdDev) {
				fmt.Printf("%s: Out of spec %s process with %s\n", newRun[i].Kind, newRun[i].Results[j].Kind, newRun[i].Results[j].Resource)
				return
			}
		}
	}
}

func readResultJSON(file string) ([]result.Host, error) {
	var res []result.Host
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(raw, &res)
	return res, nil
}

func areResultsSimilar(old, new []result.Host) bool {
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
