package file_system

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var fileCount = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "sequoia_file_count",
	Help: "The number of files on disk",
})
