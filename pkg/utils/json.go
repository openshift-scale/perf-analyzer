package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path"
	"regexp"

	"github.com/openshift/origin/test/extended/cluster/metrics"
	"github.com/redhat-performance/pbench-analyzer/pkg/result"
)

// WriteJSON will output all the calculated results to JSON file
func WriteJSON(resultDir string, r result.Result) error {
	// Remove NaN as encoding/json doesn't support NaN
	r.Hosts = removeNaN(r.Hosts)

	// Serialize results as JSON
	outHosts, err := json.Marshal(r)
	if err != nil {
		return err
	}
	// Write serialized bytes to disk
	err = ioutil.WriteFile(resultDir+"out.json", outHosts, 0644)
	if err != nil {
		return err
	}
	return nil
}

func removeNaN(hosts []result.Host) []result.Host {
	var newHosts []result.Host
	for i := range hosts {
		newHosts = append(newHosts, result.Host{})
		newHosts[i] = result.Host{
			Kind: hosts[i].Kind,
		}
		for j, result := range hosts[i].Results {
			if !math.IsNaN(result.Avg) && !math.IsNaN(result.Max) &&
				!math.IsNaN(result.Min) && !math.IsNaN(result.Pct95) {
				newHosts[i].Results = append(newHosts[i].Results, hosts[i].Results[j])
			}
		}
	}
	return newHosts
}

func GetMetrics(searchDir string, m *[]metrics.Metrics) error {
	resultFilePath := path.Join(path.Dir(path.Clean(searchDir)), "result.txt")

	bytes, err := ioutil.ReadFile(resultFilePath)
	if err != nil {
		return err
	}

	// any line start with '{' and and with '}'
	r := regexp.MustCompile(`(?m:^{.*}$)`)

	var bm metrics.BaseMetrics
	for _, jsonBytes := range r.FindAll(bytes, -1) {
		err := json.Unmarshal(jsonBytes, &bm)
		if err != nil {
			fmt.Printf("cannot unmarshal the line '%s' for BaseMetrics: %v\n", jsonBytes, err)
		}

		switch bm.Type {
		case "metrics.TestDuration":
			var td metrics.TestDuration
			err := json.Unmarshal(jsonBytes, &td)
			if err != nil {
				fmt.Printf("cannot unmarshal the line '%s' for TestDuration: %v\n", jsonBytes, err)
			}
			*m = append(*m, td)
		default:
			fmt.Printf("unsupported metrics type %v in line: %s\n", bm.Type, jsonBytes)
		}
	}

	if len(*m) == 0 {
		return fmt.Errorf("cannot find metrics in file: %s\n", resultFilePath)
	}

	return nil
}
