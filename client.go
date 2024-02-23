package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	totalDuration *prometheus.HistogramVec
	totalErrors   *prometheus.CounterVec
	httpClient    http.Client
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	buckets := []float64{.0005, .001, .005, .01, .025, .05, .1, .25, .5}
	m := &metrics{
		totalDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "latency_exporter_seconds_total",
			Help:    "Histogram of total duration.",
			Buckets: buckets,
		}, []string{"peer"}),
		totalErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "latency_exporter_errors_total",
			Help: "Number of total errors",
		}, []string{"peer"}),
		httpClient: http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
			},
			Timeout: 30 * time.Second,
		},
	}
	reg.MustRegister(m.totalDuration)
	reg.MustRegister(m.totalErrors)
	return m
}
func (m *metrics) run(peers []string) {
	time.Sleep(5 * time.Second)

	for {
		for _, peer := range peers {
			err := m.measureRequest(peer)
			if err != nil {
				m.totalErrors.WithLabelValues(peer).Inc()
				slog.Error("error during measurement", "error", err)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (m *metrics) measureRequest(peer string) error {

	timeStart := time.Now()
	req, err := http.NewRequest("GET", "http://"+peer+"/ping", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Connection", "keep-alive")

	res, err := m.httpClient.Do(req)
	timeDone := time.Now()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error response: %d", res.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	_, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %w", err)
	}
	res.Body.Close()
	m.totalDuration.WithLabelValues(peer).Observe(float64(timeDone.Sub(timeStart).Seconds()))

	return nil
}
