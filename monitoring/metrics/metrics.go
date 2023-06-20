// Package metrics sets up and handles our promethous collectors.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	messagesReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_received",
		Help: "Number of messages received from PubSub.",
	},
		[]string{"message_type"},
	)
	errorsOccurred = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_occurred",
		Help: "Number of errors occurred when processing PubSub messages.",
	},
		[]string{"processed_message_type"},
	)
	timeToProcess = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "task_duration",
		Help:    "Amount of time spent processing.",
		Buckets: []float64{.001, .002, .003, .004, .005, .01, .02, .03, .04, .05, .1, .2, .3, .4, .5},
	})
)

// RegisterPrometheusCollectors tells prometheus to set up collectors.
func RegisterPrometheusCollectors() {
	prometheus.MustRegister(messagesReceived, errorsOccurred, timeToProcess)
}

// ReceivedMessage records number of messages of each type received.
func ReceivedMessage(msgType string, t float64) {
	messagesReceived.WithLabelValues(msgType).Add(t)
}

// OccurredError records number of errors occured while processing messages
// of each type.
func OccurredError(msgType string) {
	errorsOccurred.WithLabelValues(msgType).Add(1)
}

// ObserveTimeToProcess records amount of time spent processing messages.
func ObserveTimeToProcess(t float64) {
	timeToProcess.Observe(t)
}
