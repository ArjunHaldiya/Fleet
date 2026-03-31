package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	pg  *pgxpool.Pool
	rdb *redis.Client
}

func newStorage() *Storage {
	ctx := context.Background()
	pg, err := pgxpool.New(ctx, "postgres://fleet:fleet@localhost:5432/fleetpulse?sslmode=disable")
	if err != nil {
		log.Fatal("could not connect to postgres", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("could not connect to redis:", err)
	}

	log.Println("Connected to TimeScaleDb and Redis")
	return &Storage{pg: pg, rdb: rdb}
}

func (s *Storage) save(event TelemetryEvent) error {
	ctx := context.Background()

	_, err := s.pg.Exec(ctx,
		`INSERT INTO telemetry_events (vehicle_id, speed, latitude, longitude, event_time)
	VALUES ($1, $2, $3, $4, $5)`,
		event.VehicleID, event.Speed, event.Latitude, event.Longitude, event.Timestamp)

	if err != nil {
		return fmt.Errorf("postgres write failed: %w", err)
	}

	data, _ := json.Marshal(event)
	key := fmt.Sprintf("vehicle:%s:state", event.VehicleID)
	s.rdb.Set(ctx, key, data, 30*time.Second)

	return nil
}

func (s *Storage) detectAnomaly(event TelemetryEvent) bool {
	ctx := context.Background()
	key := fmt.Sprintf("vehicle:%s:state", event.VehicleID)
	val, err := s.rdb.Get(ctx, key).Result()

	if err != nil {
		return false
	}
	var last TelemetryEvent
	if err := json.Unmarshal([]byte(val), &last); err != nil {
		return false
	}

	delta := math.Abs(event.Speed - last.Speed)

	return delta > 30
}
