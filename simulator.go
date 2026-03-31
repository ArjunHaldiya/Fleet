package main

import (
	"math/rand"
	"sync"
	"time"
)

type TelemetryEvent struct {
	VehicleID string    `json:"vehicle_id"`
	Speed     float64   `json:"speed"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

func runVehicle(id string, ch chan TelemetryEvent, wg *sync.WaitGroup) {
	defer wg.Done()
	activeVehicles.Inc()
	defer activeVehicles.Dec()

	for i := 0; i < 50; i++ {
		speed := rand.Float64()*40 + 50
		if i%10 == 0 {
			speed = 20 + rand.Float64()*10 // sudden drop to 20-30 mph — guaranteed anomaly
		}
		event := TelemetryEvent{
			VehicleID: id,
			Speed:     speed,
			Latitude:  37.4419 + rand.Float64()*0.01,
			Longitude: -122.1430 + rand.Float64()*0.01,
			Timestamp: time.Now(),
		}
		ch <- event
		time.Sleep(300 * time.Millisecond)

	}

}
