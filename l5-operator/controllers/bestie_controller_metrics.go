package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	applicationUpgradeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "bestie_upgrade_counter",
			Help: "Number of successfull bestie application upgrades processed",
		},
	)
	applicationUpgradeFailuresCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "bestie_upgrade_failures_counter",
			Help: "Number of failed bestie application upgrades",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(applicationUpgradeCounter, applicationUpgradeFailuresCounter)
}
