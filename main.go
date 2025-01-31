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
	peerFlag := flag.String("peers", "127.0.0.1:9080", "peers")
	flag.Parse()
	peers := strings.Split(*peerFlag, ",")

	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	m := NewMetrics(reg, peers)
	m.init()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_,err:=w.Write([]byte("pong"))
		if err!=nil{
			slog.Error("error writing response", "error", err)
		}
	})

	go m.run()
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
