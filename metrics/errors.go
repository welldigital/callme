package metrics

import "github.com/prometheus/client_golang/prometheus"

// ErrorCounts is a metric for the count of errors encountered by the process.
var ErrorCounts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "logging",
		Name:      "error_total",
		Help:      "The count of errors logged.",
	})
