package utils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/redhat-performance/pbench-analyzer/pkg/result"
)

// WriteCSV will write the result data to a CSV file
func WriteCSV(resultDir string, keys []string, fileHeader map[string][]string, hosts []result.Host) error {
	csvFile, err := os.Create(resultDir + "out.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Write test CSV data to stdout
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Create header & write
	header := createHeaders(keys, fileHeader)
	for _, h := range header {
		writer.Write(h)
	}

	// TODO: Maybe use reflection to get these fields instead
	stats := []string{"min", "mean", "p95", "max"}
	// Write all stats
	for _, v := range stats {
		writer.Write([]string{v})
		// Write result dataset
		for i := range hosts {
			writer.Write(hosts[i].ToSlice(v))
		}
	}
	return nil
}

// ReadCSV will return a 2d slice containing the CSV data
func ReadCSV(file string) ([][]string, error) {
	fmt.Println(file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	result, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createHeaders(keys []string, fileHeader map[string][]string) (header [][]string) {
	empty := []string{""}
	header = append(header, empty)
	header = append(header, empty)
	for _, key := range keys {
		// header keys are filenames, so we want to truncate the extension
		k := strings.Split(key, ".")
		for i := 0; i < len(fileHeader[key]); i++ {
			header[0] = append(header[0], k[0])
		}
		for _, head := range fileHeader[key] {
			header[1] = append(header[1], cleanWord(head))

		}
	}
	return
}

func cleanWord(dirty string) string {
	reg := regexp.MustCompile(`[^\w|-]+`)
	return reg.ReplaceAllString(dirty, "")
}
