package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsIngested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "events_ingested_total",
		Help: "Total number of telemetry events successfully ingested",
	})

	eventsRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "events_rejected_total",
		Help: "Total number of telemetry events rejected",
	})

	ingestLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ingest_latency_seconds",
		Help:    "Time taken to process each ingest request",
		Buckets: prometheus.DefBuckets,
	})

	activeVehicles = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_vehicles",
		Help: "Number of vehicles currently active in the simulator",
	})

	anomaliesDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "anomalies_detected_total",
		Help: "Telemetry events flagged as anomalous",
	})

	highPriorityEvents = promauto.NewCounter(prometheus.CounterOpts{
		Name: "high_priority_events_total",
		Help: "Critical telemetry events processed with priority",
	})

	eventsAgeSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "event_age_seconds",
		Help:    "Age of telemetry event from vehicle timestamp to ingest",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
	})
)
