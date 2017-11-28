package metrics

import "github.com/prometheus/client_golang/prometheus"

// JobLeaseCounts is a metric for the count of database calls made to lease jobs.
var JobLeaseCounts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "jobworker",
		Name:      "job_leased_total",
		Help:      "The count of database calls to lease jobs.",
	},
	[]string{"status"},
)

// JobLeaseDurations is a metric for the time taken to lease jobs.
var JobLeaseDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "jobworker",
	Name:      "job_leased_duration_milliseconds",
	Help:      "Time taken to lease jobs",
}, []string{"status"})

// JobExecutedCounts is a metric for the number of jobs executed.
var JobExecutedCounts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "jobworker",
		Name:      "job_executed_total",
		Help:      "The count of jobs executed.",
	},
	[]string{"status"},
)

// JobExecutedDurations is a metric for the time taken to execute jobs.
var JobExecutedDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "jobworker",
	Name:      "job_executed_duration_milliseconds",
	Help:      "Time taken to execute jobs.",
}, []string{"status"})

// JobExecutedDelay is a metric for the delay between when jobs should have run, and when they did.
var JobExecutedDelay = prometheus.NewHistogram(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "jobworker",
	Name:      "job_executed_delay_milliseconds",
	Help:      "The delay between when jobs should have run, and when they did.",
})

// JobCompletedCounts is a metric for the number of jobs marked as completed.
var JobCompletedCounts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "jobworker",
		Name:      "job_completed_total",
		Help:      "The count of jobs marked as completed.",
	},
	[]string{"status"},
)

// JobCompletedDurations is a metric for the time taken to mark jobs as completed.
var JobCompletedDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "jobworker",
	Name:      "job_completed_duration_milliseconds",
	Help:      "Time taken to mark jobs as completed.",
}, []string{"status"})
