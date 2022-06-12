package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	prometheus.MustRegister(pingCounter)

	http.HandleFunc("/ping", ping)
	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":8090", nil)
}

func ping(w http.ResponseWriter, req *http.Request) {
	pingCounter.Inc()
	fmt.Fprintf(w, "pong")
}

var pingCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "ping_request_count",
		Help: "No of request handled by Ping handler",
	},
)
