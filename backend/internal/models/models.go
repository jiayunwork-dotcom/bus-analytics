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

type LineHealthSubScore struct {
	Name       string  `json:"name"`
	RawValue   float64 `json:"raw_value"`
	RawValueStr string `json:"raw_value_str"`
	Score      float64 `json:"score"`
	MaxScore   float64 `json:"max_score"`
	ScoreRatio float64 `json:"score_ratio"`
}

type LineHealthScore struct {
	LineNo           string             `json:"line_no"`
	LineName         string             `json:"line_name"`
	TotalScore       float64            `json:"total_score"`
	ScoreLevel       string             `json:"score_level"`
	PassengerIntensityScore LineHealthSubScore `json:"passenger_intensity"`
	PeakLoadFactorScore     LineHealthSubScore `json:"peak_load_factor"`
	OnTimeRateScore         LineHealthSubScore `json:"on_time_rate"`
	OperatingSpeedScore     LineHealthSubScore `json:"operating_speed"`
	VehicleUtilizationScore LineHealthSubScore `json:"vehicle_utilization"`
	AvgDailyTripsPerVehicle float64 `json:"avg_daily_trips_per_vehicle"`
}

type SimParams struct {
	LineNo           string `json:"line_no" binding:"required"`
	PeakInterval     int    `json:"peak_interval"`
	OffPeakInterval  int    `json:"off_peak_interval"`
	StationDelta     int    `json:"station_delta"`
	Date             string `json:"date"`
}

type SingleLineSimParams struct {
	LineNo          string `json:"line_no" binding:"required"`
	PeakInterval    int    `json:"peak_interval"`
	OffPeakInterval int    `json:"off_peak_interval"`
	StationDelta    int    `json:"station_delta"`
}

type JointSimParams struct {
	Lines []SingleLineSimParams `json:"lines" binding:"required,min=1,max=3"`
	Date  string                `json:"date"`
}

type SimKPI struct {
	DailyTrips        int     `json:"daily_trips"`
	PeakLoadFactor    float64 `json:"peak_load_factor"`
	OperatingSpeed    float64 `json:"operating_speed"`
	PassengerIntensity float64 `json:"passenger_intensity"`
}

type AdjLineImpact struct {
	LineNo           string  `json:"line_no"`
	LineName         string  `json:"line_name"`
	OrigPeakLoad     float64 `json:"orig_peak_load"`
	NewPeakLoad      float64 `json:"new_peak_load"`
	LoadIncrement    float64 `json:"load_increment"`
	OverloadRisk     bool    `json:"overload_risk"`
	SharedStations   []string `json:"shared_stations"`
}

type TrendPoint struct {
	RemoveCount   int     `json:"remove_count"`
	PeakLoadFactor float64 `json:"peak_load_factor"`
}

type SimResult struct {
	LineNo              string          `json:"line_no"`
	LineName            string          `json:"line_name"`
	OrigStationCount    int             `json:"orig_station_count"`
	NewStationCount     int             `json:"new_station_count"`
	OrigPeakInterval    int             `json:"orig_peak_interval"`
	NewPeakInterval     int             `json:"new_peak_interval"`
	OrigOffPeakInterval int             `json:"orig_off_peak_interval"`
	NewOffPeakInterval  int             `json:"new_off_peak_interval"`
	OrigTotalTrips      int             `json:"orig_total_trips"`
	NewTotalTrips       int             `json:"new_total_trips"`
	TripsDelta          int             `json:"trips_delta"`
	PeakCapacityChange  float64         `json:"peak_capacity_change"`
	CapacityWarning     bool            `json:"capacity_warning"`
	AvailableVehicles   int             `json:"available_vehicles"`
	OrigKPI             SimKPI          `json:"orig_kpi"`
	NewKPI              SimKPI          `json:"new_kpi"`
	AdjacentImpacts     []AdjLineImpact `json:"adjacent_impacts"`
	RemovalTrend        []TrendPoint    `json:"removal_trend"`
	OrigPeakTrips       int             `json:"orig_peak_trips"`
	NewPeakTrips        int             `json:"new_peak_trips"`
}

type SharedVehicleConflict struct {
	VehicleNo       string   `json:"vehicle_no"`
	InvolvedLines   []string `json:"involved_lines"`
	TotalPeakTrips  int      `json:"total_peak_trips"`
	CapacityLimit   int      `json:"capacity_limit"`
}

type AdjLineMergedImpact struct {
	LineNo         string   `json:"line_no"`
	LineName       string   `json:"line_name"`
	OrigPeakLoad   float64  `json:"orig_peak_load"`
	NewPeakLoad    float64  `json:"new_peak_load"`
	LoadIncrement  float64  `json:"load_increment"`
	OverloadRisk   bool     `json:"overload_risk"`
	SharedStations []string `json:"shared_stations"`
	AffectedBy     []string `json:"affected_by"`
}

type JointOverview struct {
	TotalOrigTrips         int                  `json:"total_orig_trips"`
	TotalNewTrips          int                  `json:"total_new_trips"`
	TotalTripsDelta        int                  `json:"total_trips_delta"`
	TotalTripsChangePct    float64              `json:"total_trips_change_pct"`
	AvgOrigLoadFactor      float64              `json:"avg_orig_load_factor"`
	AvgNewLoadFactor       float64              `json:"avg_new_load_factor"`
	AvgLoadFactorDelta     float64              `json:"avg_load_factor_delta"`
	SharedVehicleConflicts []SharedVehicleConflict `json:"shared_vehicle_conflicts"`
	HasVehicleConflict     bool                 `json:"has_vehicle_conflict"`
	MergedAdjacentImpacts  []AdjLineMergedImpact `json:"merged_adjacent_impacts"`
	AdjacentOverloadCount  int                  `json:"adjacent_overload_count"`
}

type JointSimResult struct {
	LineResults    []SimResult   `json:"line_results"`
	JointOverview  JointOverview `json:"joint_overview"`
	LineColors     []string      `json:"line_colors"`
}

