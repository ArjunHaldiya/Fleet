package main

import (
	"fmt"
	"log"
	"sync"
)

func main() {
	store := newStorage()
	ch := make(chan TelemetryEvent, 500)
	var wg sync.WaitGroup

	go startServer(ch, store)

	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("vehicle-%03d", i)
		wg.Add(1)
		go runVehicle(id, ch, &wg)
	}

	go func() {
		for event := range ch {
			fmt.Printf("[%s] speed=%.1f lat=%.4f\n",
				event.VehicleID, event.Speed, event.Latitude)
			if err := store.save(event); err != nil {
				log.Println("Save error", err)
			}
			eventsIngested.Inc()
		}
	}()
	wg.Wait()
	fmt.Printf("Simulator done, Server running")

	select {}
}
