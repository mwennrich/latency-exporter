package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type connDetails struct {
	addr net.Addr
}

type metrics struct {
	totalDuration         *prometheus.HistogramVec
	totalErrors           *prometheus.CounterVec
	totalConnectionResets *prometheus.CounterVec
	httpClients           map[string]*http.Client
	connDetails           map[string]*connDetails
	peers                 []string
}

func NewMetrics(reg prometheus.Registerer, peers []string) *metrics {
	buckets := []float64{.005, .01, .015, .02, .025, .05, .1, .25, .5, 1}

	httpClients := map[string]*http.Client{}
	cds := map[string]*connDetails{}
	for _, p := range peers {
		cds[p] = &connDetails{}

		httpClients[p] = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
				DialContext:         cds[p].DialContext,
			},
			Timeout: 30 * time.Second,
		}
	}

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

		totalConnectionResets: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "latency_exporter_connection_resets_total",
			Help: "Number of total connection resets",
		}, []string{"peer"}),

		httpClients: httpClients,
		connDetails: cds,
		peers:       peers,
	}

	reg.MustRegister(m.totalDuration)
	reg.MustRegister(m.totalErrors)
	reg.MustRegister(m.totalConnectionResets)

	return m
}
func (m *metrics) run() {
	localAdresses := map[string]string{}

	time.Sleep(5 * time.Second)

	for {
		for _, peer := range m.peers {
			la, err := m.measureRequest(peer)
			if _, ok := localAdresses[peer]; !ok {
				localAdresses[peer] = la
			} else if localAdresses[peer] != la {
				m.totalConnectionResets.WithLabelValues(peer).Inc()
				localAdresses[peer] = la
			}
			if err != nil {
				m.totalErrors.WithLabelValues(peer).Inc()
				slog.Error("error during measurement", "error", err)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (m *metrics) measureRequest(peer string) (string, error) {

	timeStart := time.Now()
	req, err := http.NewRequest("GET", "http://"+peer+"/ping", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Connection", "keep-alive")

	res, err := m.httpClients[peer].Do(req)
	timeDone := time.Now()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response: %d", res.StatusCode)
	}
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	_, err = io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %w", err)
	}
	res.Body.Close()
	m.totalDuration.WithLabelValues(peer).Observe(float64(timeDone.Sub(timeStart).Seconds()))

	// fmt.Println("Local address:", m.connDetails[peer].addr.String())

	return m.connDetails[peer].addr.String(), nil
}

func (m *metrics) init() {
	for _, peer := range m.peers {
		m.totalErrors.WithLabelValues(peer).Add(0)
		m.totalConnectionResets.WithLabelValues(peer).Add(0)
	}
}
func (cd *connDetails) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	cd.addr = conn.LocalAddr()

	return conn, nil
}
