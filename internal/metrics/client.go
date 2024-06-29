package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"time"
)

var requestHttpMetrics = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  "sso",
	Subsystem:  "http",
	Name:       "request",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
}, []string{"status"})

var requestGRPCMetrics = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace:  "sso",
	Subsystem:  "grpc",
	Name:       "request",
	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
}, []string{"status"})

func ObserveHttpRequest(d time.Duration, status int) {
	requestHttpMetrics.WithLabelValues(strconv.Itoa(status)).Observe(d.Seconds())
}

func ObserveGRPCRequest(d time.Duration, status int) {
	requestGRPCMetrics.WithLabelValues(strconv.Itoa(status)).Observe(d.Seconds())
}
