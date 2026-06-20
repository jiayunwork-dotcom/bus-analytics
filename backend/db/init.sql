CREATE TABLE IF NOT EXISTS routes (
    id SERIAL PRIMARY KEY,
    line_no VARCHAR(50) NOT NULL UNIQUE,
    line_name VARCHAR(200) NOT NULL,
    start_station VARCHAR(200) NOT NULL,
    end_station VARCHAR(200) NOT NULL,
    total_km DECIMAL(10,2) NOT NULL,
    station_count INTEGER NOT NULL,
    fare DECIMAL(10,2) NOT NULL,
    straight_line_distance_km DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS trips (
    id SERIAL PRIMARY KEY,
    line_no VARCHAR(50) NOT NULL,
    trip_date DATE NOT NULL,
    trip_no VARCHAR(100) NOT NULL,
    actual_departure_time TIME NOT NULL,
    planned_departure_time TIME NOT NULL,
    vehicle_no VARCHAR(50) NOT NULL,
    driver_name VARCHAR(100),
    direction SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS station_flows (
    id SERIAL PRIMARY KEY,
    line_no VARCHAR(50) NOT NULL,
    flow_date DATE NOT NULL,
    trip_no VARCHAR(100) NOT NULL,
    station_seq INTEGER NOT NULL,
    station_name VARCHAR(200) NOT NULL,
    board_count INTEGER NOT NULL DEFAULT 0,
    alight_count INTEGER NOT NULL DEFAULT 0,
    card_count INTEGER NOT NULL DEFAULT 0,
    card_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vehicle_mileages (
    id SERIAL PRIMARY KEY,
    vehicle_no VARCHAR(50) NOT NULL,
    mileage_date DATE NOT NULL,
    total_km DECIMAL(10,2) NOT NULL,
    operating_km DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_trips_line_date ON trips(line_no, trip_date);
CREATE INDEX IF NOT EXISTS idx_flows_line_date ON station_flows(line_no, flow_date);
CREATE INDEX IF NOT EXISTS idx_flows_card ON station_flows(card_id, flow_date);
CREATE INDEX IF NOT EXISTS idx_vehicle_date ON vehicle_mileages(vehicle_no, mileage_date);
