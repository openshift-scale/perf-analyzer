package prometheus

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/openshift-scale/perf-analyzer/pkg/config"
	"github.com/openshift-scale/perf-analyzer/pkg/result"
	"github.com/openshift-scale/perf-analyzer/pkg/stats"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

var insecureRoundTripper http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	TLSHandshakeTimeout: 10 * time.Second,
}

type TimePoint [2]float64
type TimeSeriesPoints []TimePoint

type TimeSeries struct {
	Name   string
	Points TimeSeriesPoints
	Tags   []rawData
}

type rawData struct {
	Name   string
	Values []float64
}

type prometheusConfig struct {
	config api.Config
	API    v1.API
	Errors []error
}

func newPrometheusConfig(url, token string, insecureTLS bool) (prometheusConfig, error) {
	rt, err := NewBearerAuthRoundTripper(token, insecureRoundTripper)
	if err != nil {
		return prometheusConfig{}, err
	}

	return prometheusConfig{config: api.Config{Address: url, RoundTripper: rt}}, nil
}

func NewBearerAuthRoundTripper(bearer string, rt http.RoundTripper) (http.RoundTripper, error) {
	if len(bearer) == 0 {
		return nil, fmt.Errorf("No bearer token provided for RoundTripper\n")
	}
	return &bearerAuthRoundTripper{bearer, rt}, nil
}

type bearerAuthRoundTripper struct {
	bearer string
	rt     http.RoundTripper
}

func (rt *bearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.rt.RoundTrip(req)
	}

	token := rt.bearer
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return rt.rt.RoundTrip(req)
}

func (c *prometheusConfig) newPrometheusAPI() {
	client, err := api.NewClient(c.config)
	if err != nil {
		fmt.Printf("Error creating Prometheus client: %v\n", err)
		c.Errors = append(c.Errors, err)
	} else {
		c.API = v1.NewAPI(client)
	}
}

// DoPrometheusQuery will run queries against Prometheus endpoint
func DoPrometheusQuery(cfg config.ScrapeConfig) {
	config, err := newPrometheusConfig(cfg.UrlFlag, cfg.TokenFlag, cfg.InsecureTLSFlag)
	if err != nil {
		fmt.Printf("Unable to create Prometheus config: %v\n", err)
	}
	config.newPrometheusAPI()
	if len(config.Errors) != 0 {
		fmt.Printf("Client Error %v\n", config.Errors[0])
	}

	queries := map[string]string{
		"cpu":    `sum(rate(container_cpu_usage_seconds_total{job="kubelet", image!="", container!="POD"}[5m])) by (namespace)`,
		"memory": `sum(container_memory_usage_bytes{container_name!=""}) by (namespace)`,
	}
	end := time.Now()
	start := end.Add(time.Duration(-1*cfg.DurationFlag) * time.Minute)
	step, err := time.ParseDuration("1m")
	if err != nil {
		fmt.Printf("Error parsing step duration %s\n", cfg.StepFlag)
	}
	r := v1.Range{Start: start, End: end, Step: step}

	var results []result.ResultType
	for resource, query := range queries {
		queryResult, _, err := config.API.QueryRange(context.Background(), query, r)
		if err != nil {
			fmt.Printf("Prometheus query error: %v\n", err)
			return
		}
		fmt.Printf("Query returned: %+v\n", queryResult)

		data, ok := queryResult.(model.Matrix)
		if !ok {
			fmt.Printf("Unsupported result format: %s\n", queryResult.Type().String())
			return
		}

		series := TimeSeries{
			Name: query,
			Tags: []rawData{},
		}

		for _, j := range data {
			var newData rawData
			for _, v := range j.Metric {
				newData.Name = string(v)
			}
			for _, v := range j.Values {
				newData.Values = append(newData.Values, float64(v.Value))
			}
			series.Tags = append(series.Tags, newData)
		}
		fmt.Fprintf(os.Stderr, "Series: %+v\n", series)

		for _, r := range series.Tags {
			results = append(results, AddResult(r.Values, r.Name))
			results[len(results)-1].Resource = resource
		}
	}

	fmt.Printf("Results: %+v\n", results)

}

func AddResult(newResult []float64, kind string) result.ResultType {
	min, _ := stats.Minimum(newResult)
	max, _ := stats.Maximum(newResult)
	avg, _ := stats.Mean(newResult)
	pct95, _ := stats.Percentile(newResult, 95)

	result := result.ResultType{
		Kind:  kind,
		Min:   min,
		Max:   max,
		Avg:   avg,
		Pct95: pct95,
	}

	return result

}
