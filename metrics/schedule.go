package metrics

import "github.com/prometheus/client_golang/prometheus"

// ScheduleLeaseCounts is a metric for the count of database calls made to lease schedules.
var ScheduleLeaseCounts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "scheduleworker",
		Name:      "schedule_leased_total",
		Help:      "The count of database calls to lease schedules.",
	},
	[]string{"status"},
)

// ScheduleLeaseDurations is a metric for the time taken to lease schedules.
var ScheduleLeaseDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "scheduleworker",
	Name:      "schedule_leased_duration_milliseconds",
	Help:      "Time taken to lease schedules.",
}, []string{"status"})

// ScheduleExecutedCounts is a metric for the number of schedules executed.
var ScheduleExecutedCounts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "scheduleworker",
		Name:      "schedule_executed_total",
		Help:      "The count of schedules executed.",
	},
	[]string{"status"},
)

// ScheduleExecutedDelay is a metric for the delay between when schedules should have run, and when they did.
var ScheduleExecutedDelay = prometheus.NewHistogram(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "scheduleworker",
	Name:      "schedule_executed_delay_milliseconds",
	Help:      "The delay between when schedules should have run, and when they did.",
})

// ScheduleJobStartedCounts is a metric for the count of jobs started by a schedule.
var ScheduleJobStartedCounts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "callme",
		Subsystem: "scheduleworker",
		Name:      "schedule_job_started_total",
		Help:      "The count of jobs started by a schedule.",
	},
	[]string{"status"},
)

// ScheduleJobStartedDurations is a metric for the time taken to start jobs based on a schedule.
var ScheduleJobStartedDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "callme",
	Subsystem: "scheduleworker",
	Name:      "schedule_job_started_duration_milliseconds",
	Help:      "Time taken to start jobs based on a schedule.",
}, []string{"status"})
