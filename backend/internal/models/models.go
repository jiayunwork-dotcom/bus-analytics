package models

import "time"

type Route struct {
	ID                     int       `json:"id"`
	LineNo                 string    `json:"line_no"`
	LineName               string    `json:"line_name"`
	StartStation           string    `json:"start_station"`
	EndStation             string    `json:"end_station"`
	TotalKm                float64   `json:"total_km"`
	StationCount           int       `json:"station_count"`
	Fare                   float64   `json:"fare"`
	StraightLineDistanceKm *float64  `json:"straight_line_distance_km,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}

type Trip struct {
	ID                   int       `json:"id"`
	LineNo               string    `json:"line_no"`
	TripDate             string    `json:"trip_date"`
	TripNo               string    `json:"trip_no"`
	ActualDepartureTime  string    `json:"actual_departure_time"`
	PlannedDepartureTime string    `json:"planned_departure_time"`
	VehicleNo            string    `json:"vehicle_no"`
	DriverName           string    `json:"driver_name"`
	Direction            int       `json:"direction"`
	CreatedAt            time.Time `json:"created_at"`
}

type StationFlow struct {
	ID          int       `json:"id"`
	LineNo      string    `json:"line_no"`
	FlowDate    string    `json:"flow_date"`
	TripNo      string    `json:"trip_no"`
	StationSeq  int       `json:"station_seq"`
	StationName string    `json:"station_name"`
	BoardCount  int       `json:"board_count"`
	AlightCount int       `json:"alight_count"`
	CardCount   int       `json:"card_count"`
	CardID      string    `json:"card_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type VehicleMileage struct {
	ID          int       `json:"id"`
	VehicleNo   string    `json:"vehicle_no"`
	MileageDate string    `json:"mileage_date"`
	TotalKm     float64   `json:"total_km"`
	OperatingKm float64   `json:"operating_km"`
	CreatedAt   time.Time `json:"created_at"`
}

type DataSummary struct {
	DateRange   []string `json:"date_range"`
	LineCount   int      `json:"line_count"`
	TripCount   int      `json:"trip_count"`
	VehicleCount int     `json:"vehicle_count"`
	TotalPassengers int64 `json:"total_passengers"`
}

type LineEfficiency struct {
	LineNo            string  `json:"line_no"`
	LineName          string  `json:"line_name"`
	PassengerIntensity float64 `json:"passenger_intensity"`
	PeakLoadFactor    float64 `json:"peak_load_factor"`
	OffPeakLoadFactor float64 `json:"off_peak_load_factor"`
	OperatingSpeed    float64 `json:"operating_speed"`
	OnTimeRate        float64 `json:"on_time_rate"`
	Level             string  `json:"level"`
}

type SectionFlow struct {
	StationSeq     int     `json:"station_seq"`
	FromStation    string  `json:"from_station"`
	ToStation      string  `json:"to_station"`
	Passengers     int     `json:"passengers"`
	Direction      int     `json:"direction"`
}

type HourlyDistribution struct {
	Hour     int `json:"hour"`
	Boarders int `json:"boarders"`
}

type ODPair struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Count       int    `json:"count"`
}

type ScheduleOptimization struct {
	PeriodName            string  `json:"period_name"`
	TimeRange             string  `json:"time_range"`
	CurrentInterval       int     `json:"current_interval"`
	RecommendedInterval   int     `json:"recommended_interval"`
	RecommendedTrips      int     `json:"recommended_trips"`
	IsPeak                bool    `json:"is_peak"`
}

type ScheduleResult struct {
	Optimizations        []ScheduleOptimization `json:"optimizations"`
	TotalTrips           int                    `json:"total_trips"`
	CurrentAvgWaitTime   float64                `json:"current_avg_wait_time"`
	OptimizedAvgWaitTime float64                `json:"optimized_avg_wait_time"`
	ImprovementPercent   float64                `json:"improvement_percent"`
}

type VehicleUtilization struct {
	VehicleNo     string  `json:"vehicle_no"`
	DailyTrips    int     `json:"daily_trips"`
	FirstTripTime string  `json:"first_trip_time"`
	LastTripTime  string  `json:"last_trip_time"`
	OperatingKm   float64 `json:"operating_km"`
	Level         string  `json:"level"`
}

type NetworkMetrics struct {
	StationCoverage   float64 `json:"station_coverage"`
	LineDuplication   float64 `json:"line_duplication"`
	NonStraightLine   float64 `json:"non_straight_line"`
	StationCount      int     `json:"station_count"`
	CityArea          float64 `json:"city_area"`
	TotalLineKm       float64 `json:"total_line_km"`
	UniqueLineKm      float64 `json:"unique_line_km"`
}

type ValidationError struct {
	Row     int      `json:"row"`
	Fields  []string `json:"fields"`
}

type UploadResponse struct {
	Success       bool              `json:"success"`
	Message       string            `json:"message"`
	DataType      string            `json:"data_type"`
	InsertedCount int               `json:"inserted_count"`
	Errors        []ValidationError `json:"errors"`
}
