package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	store := newStorage()
	producer := newProducer()
	defer producer.Close()

	ch:= make(chan TelemetryEvent, 500)
	var wg sync.WaitGroup

	go startServer(ch, store)
	time.Sleep(100 * time.Millisecond)

	go startConsumer(store)
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("vehicle-%03d", i)
		wg.Add(1)
		go runVehicle(id, producer, &wg)
	}

	go func() {
		for event := range ch {
			age := time.Since(event.Timestamp).Seconds()
			eventsAgeSeconds.Observe(age)

			fmt.Printf("[%s] speed=%.1f priority=%s\n",
				event.VehicleID, event.Speed, event.Priority)
			if event.Priority == "high" {
				if err := store.save(event); err != nil {
					log.Println("priority save error: ", err)
				}
				highPriorityEvents.Inc()
				log.Printf("HIGH PRIORITY: vehicle=%s speed=%.1f",
					event.VehicleID, event.Speed)
			} else {
				if store.detectAnomaly(event) {
					anomaliesDetected.Inc()
					log.Printf("ANOMALY DETECTED: vehicle=%s speed %.1f",
						event.VehicleID, event.Speed)
				}
				if err := store.save(event); err != nil {
					log.Println("Save error", err)
				}
			}
			eventsIngested.Inc()
		}
	}()
	wg.Wait()
	fmt.Printf("Simulator done, Server running")

	select {}
}
