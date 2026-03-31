package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
	http.HandleFunc("/vehicle/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "invalid path", http.StatusBadRequest)
		}
		id := parts[2]

		rows, err := store.pg.Query(context.Background(),
			`SELECT vehicle_id, speed, latitude, longitude, event_time
			FROM telemetry_events
			WHERE vehicle_id = $1
			AND event_time > Now() - INTERVAL '5 minutes'
			ORDER BY event_time DESC`, id,
		)
		if err != nil {
			log.Printf("history query error: %v", err)
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var events []TelemetryEvent
		for rows.Next() {
			var e TelemetryEvent
			if err := rows.Scan(&e.VehicleID, &e.Speed, &e.Latitude, &e.Longitude, &e.Timestamp); err != nil {
				continue
			}
			events = append(events, e)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		health := map[string]string{
			"status":    "ok",
			"postgress": "ok",
			"redis":     "ok",
		}

		if err := store.pg.Ping(ctx); err != nil {
			health["postgres"] = "unreachable"
			health["status"] = "degraded"
		}

		if err := store.rdb.Set(ctx, "health:check", "1", 5*time.Second).Err(); err != nil {
			health["redis"] = "unreachable"
			health["status"] = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		if health["status"] == "degraded" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(health)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
