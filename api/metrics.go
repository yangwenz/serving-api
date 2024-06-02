package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var totalRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of requests",
	},
	[]string{"path"},
)

var responseStatus = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"path", "status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests",
}, []string{"path"})

/*
func init() {
	if err := prometheus.Register(totalRequests); err != nil {
		log.Fatal().Msgf("failed to register prometheus metric totalRequests: %v", err)
	}
	if err := prometheus.Register(responseStatus); err != nil {
		log.Fatal().Msgf("failed to register prometheus metric responseStatus: %v", err)
	}
	if err := prometheus.Register(httpDuration); err != nil {
		log.Fatal().Msgf("failed to register prometheus metric httpDuration: %v", err)
	}
}
*/
