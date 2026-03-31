package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func logDeadLetter(event TelemetryEvent, reason string) {
	f, err := os.OpenFile("dead_letter.json1", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Println("could not open dead letter file", err)
		return
	}
	defer f.Close()

	entry := map[string]any{
		"event":  event,
		"reason": reason,
		"time":   time.Now(),
	}

	json.NewEncoder(f).Encode(entry)
}

func startServer(ch chan TelemetryEvent, store *Storage) {
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event TelemetryEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			logDeadLetter(event, "invalid json")
			eventsRejected.Inc()
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if event.VehicleID == "" || event.Speed <= 0 {
			logDeadLetter(event, "missing required fields")
			eventsRejected.Inc()
			http.Error(w, "missing vehicle_id or speed", http.StatusBadRequest)
			return
		}

		event.Timestamp = time.Now()
		ch <- event

		if err := store.save(event); err != nil {
			log.Println("Storage error: ", err)
		}
		eventsIngested.Inc()
		ingestLatency.Observe(time.Since(start).Seconds())

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
