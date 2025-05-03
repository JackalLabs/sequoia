package queue

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var queueSize = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "sequoia_queue_size",
	Help: "The number of messages currently in the queue",
})

var _ = queueSize
