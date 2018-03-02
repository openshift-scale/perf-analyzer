package utils

import (
	"encoding/json"
	"io/ioutil"
	"math"

	"github.com/redhat-performance/pbench-analyzer/pkg/result"
)

// WriteJSON will output all the calculated results to JSON file
func WriteJSON(resultDir string, hosts []result.Host) error {
	// Remove NaN as encoding/json doesn't support NaN
	newHosts := removeNaN(hosts)

	// Serialize results as JSON
	outHosts, err := json.Marshal(newHosts)
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
