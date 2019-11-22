package main

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Resolution int64 = 10000

	metrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tmp_file_age",
		Help: "Age of /tmp/testfile",
	})
)

func GetCollector() prometheus.Collector {
	return metrics
}

func UpdateMetric() {
	stat, err := os.Stat("/tmp/testfile")
	if err == nil {
		metrics.Set(float64(time.Since(stat.ModTime()).Seconds()))
	} else {
		metrics.Set(float64(0))
	}
}
