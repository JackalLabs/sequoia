package proofs

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var filesProving = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "sequoia_current_proofs_processing",
	Help: "The number of files currently being proven",
})
