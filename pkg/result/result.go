package result

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/openshift/origin/test/extended/cluster/metrics"
	"github.com/redhat-performance/pbench-analyzer/pkg/stats"
)

type Result struct {
	Hosts   []Host
	Metrics []metrics.Metrics
}

// Host struct of a Kind has a ResultDir and a list of Results
type Host struct {
	Kind      string
	ResultDir string `json:",omitempty"`
	Results   []ResultType
}

// ResultType is a single Result summary
type ResultType struct {
	Kind                 string
	Path                 string `json:"-"`
	Resource             string
	Min, Max, Avg, Pct95 float64
}

// ToSlice helps us print the Host struct data to a CSV row
func (h *Host) ToSlice(stat string) (row []string) {
	row = append(row, h.Kind)
	switch stat {
	case "min":
		// append minimum
		for _, result := range h.Results {
			row = append(row, strconv.FormatFloat(result.Min, 'f', 2, 64))
		}
	case "mean":
		// appened mean
		for _, result := range h.Results {
			row = append(row, strconv.FormatFloat(result.Avg, 'f', 2, 64))
		}
	case "p95":
		// append p95
		for _, result := range h.Results {
			row = append(row, strconv.FormatFloat(result.Pct95, 'f', 2, 64))
		}
	case "max":
		// append max
		for _, result := range h.Results {
			row = append(row, strconv.FormatFloat(result.Max, 'f', 2, 64))
		}
	default:
		// do nothing
	}
	return
}

// AddResult will create a new ResultType which is added to a Host
func (h *Host) AddResult(newResult []float64, file string, kind string, res string) []ResultType {
	min, _ := stats.Minimum(newResult)
	max, _ := stats.Maximum(newResult)
	avg, _ := stats.Mean(newResult)
	pct95, _ := stats.Percentile(newResult, 95)

	h.Results = append(h.Results, ResultType{
		Kind:     kind,
		Path:     file,
		Resource: strings.TrimSuffix(res, ".csv"),
		Min:      min,
		Max:      max,
		Avg:      avg,
		Pct95:    pct95,
	})

	return h.Results

}

// NewSlice will extract a single slice of values from a CSV
func NewSlice(bigSlice [][]string, title string) ([]float64, error) {
	floatValues := make([]float64, len(bigSlice)-1)
	var column int
	for i, v := range bigSlice {
		if i == 0 {
			var err error
			column, err = stringPositionInSlice(title, v)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			continue
		}
		value, _ := strconv.ParseFloat(bigSlice[i][column], 64)
		floatValues[i-1] = value
	}
	return floatValues, nil
}

// TODO: handle duplicates or none
func stringPositionInSlice(a string, list []string) (int, error) {
	for i, v := range list {
		match, _ := regexp.MatchString(a, v)
		if match {
			return i, nil
		}
	}
	return 0, fmt.Errorf("No matching headers")
}
