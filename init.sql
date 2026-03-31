CREATE TABLE IF NOT EXISTS telemetry_events (
    vehicle_id  TEXT    NOT NULL,
    speed   DOUBLE PRECISION    NOT NULL,
    latitude    DOUBLE PRECISION    NOT NULL,
    longitude   DOUBLE PRECISION    NOT NULL,
    event_time  TIMESTAMPTZ     NOT NULL
);
SELECT create_hypertable('telemetry_events', 'event_time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS ON telemetry_events (vehicle_id, event_time DESC)
