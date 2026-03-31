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
		Name: "ingest_latency_seconds",
		Help: "Time taken to process each ingest request",
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
)