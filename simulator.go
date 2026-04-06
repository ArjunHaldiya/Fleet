package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

)

type TelemetryEvent struct {
	VehicleID string    `json:"vehicle_id"`
	Speed     float64   `json:"speed"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
	Priority	string	`json:"priority"`
}

func runVehicle(id string, p *kafka.Producer , wg *sync.WaitGroup) {
	defer wg.Done()
	activeVehicles.Inc()
	defer activeVehicles.Dec()

	for i := 0; i < 50; i++ {
		speed := rand.Float64()*40 + 50
		priority := "normal"

		if i%10 == 0 {
			speed = 20 + rand.Float64()*10 // sudden drop to 20-30 mph — guaranteed anomaly
			priority = "high"
		}
		event := TelemetryEvent{
			VehicleID: id,
			Speed:     speed,
			Latitude:  37.4419 + rand.Float64()*0.01,
			Longitude: -122.1430 + rand.Float64()*0.01,
			Timestamp: time.Now(),
			Priority: priority,
		}
		if err := publishEvent(p, event); err != nil {
			log.Printf("Failed to publish event: %v", err)
		}
		time.Sleep(300 * time.Millisecond)

	}

}
