package main

import (
	"net/http"
	"strings"
	"time"

	"log/slog"

	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	m := NewMetrics(reg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong\n"))
	})

	peerFlag := flag.String("peers", "127.0.0.1:9080", "peers")
	flag.Parse()
	peers := strings.Split(*peerFlag, ",")
	go m.run(peers)
	http.Handle("/ping", handler)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	slog.Info("Beginning to serve on port :9080")
	server := &http.Server{
		Addr:              ":9080",
		ReadHeaderTimeout: 300 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		slog.Error("ListenAndServe: ", "error", err)
	}
}
