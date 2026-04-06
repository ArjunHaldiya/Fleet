package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	kafkaBroker = "localhost:9092"
	kafkaTopic  = "telemetry.events"
)

func newProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
	})
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}
	log.Println("Kafka producer ready")
	return p
}

func newConsumer(groupID string) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	if err := c.SubscribeTopics([]string{kafkaTopic}, nil); err != nil {
		log.Fatal("Failed to subscribe to topic:", err)
	}
	log.Println("Kafka consumer ready, subscribed to", kafkaTopic)
	return c
}

func publishEvent(p *kafka.Producer, event TelemetryEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kafkaTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.VehicleID),
		Value: data,
	}, nil)
}

func startConsumer(store *Storage) {
	consumer := newConsumer("fleetpulse-group")
	defer consumer.Close()

	log.Println("Starting Kafka consumer loop...")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error: %v", err)
			continue
		}

		var event TelemetryEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to deserialize event: %v", err)
			continue
		}

		if store.detectAnomaly(event) {
			anomaliesDetected.Inc()
			log.Printf("ANOMALY DETECTED: vehicle=%s speed=%.1f", event.VehicleID, event.Speed)
		}

		if event.Priority == "high" {
			highPriorityEvents.Inc()
			log.Printf("HIGH PRIORITY: vehicle=%s speed=%.1f", event.VehicleID, event.Speed)
		}

		age := time.Since(event.Timestamp).Seconds()
		eventsAgeSeconds.Observe(age)

		if err := store.save(event); err != nil {
			log.Printf("Save error: %v", err)
		}

		eventsIngested.Inc()
	}
}
