package main

import (
	"math/rand"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Resolution int64 = 10000

	HddFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hd_errors_total",
			Help: "Number of hard-disk errors.",
		},
		[]string{"device"},
	)
)

func GetCollector() prometheus.Collector {
	return HddFailures
}

func UpdateMetric() {
	HddFailures.With(prometheus.Labels{"device": "/dev/sda"}).Add((float64(rand.Intn(99))))
}
