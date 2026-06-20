package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"fmt"
	"sort"
)

type VehicleService struct{}

func NewVehicleService() *VehicleService {
	return &VehicleService{}
}

type VehicleGanttItem struct {
	VehicleNo string `json:"vehicle_no"`
	LineNo    string `json:"line_no"`
	TripNo    string `json:"trip_no"`
	Direction int    `json:"direction"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	TripDate  string `json:"trip_date"`
}

func (s *VehicleService) GetVehicleUtilizations(date string) ([]models.VehicleUtilization, int, int, float64, error) {
	query := `
		SELECT 
			t.vehicle_no,
			COUNT(DISTINCT t.trip_no) as daily_trips,
			MIN(t.actual_departure_time)::text as first_trip,
			MAX(t.actual_departure_time)::text as last_trip,
			COALESCE(vm.operating_km, 0)
		FROM trips t
		LEFT JOIN vehicle_mileages vm ON t.vehicle_no = vm.vehicle_no AND t.trip_date = vm.mileage_date
		WHERE t.trip_date = $1
		GROUP BY t.vehicle_no, vm.operating_km
		ORDER BY daily_trips DESC
	`

	rows, err := db.DB.Query(query, date)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	var results []models.VehicleUtilization
	totalTrips := 0
	for rows.Next() {
		var v models.VehicleUtilization
		var firstTrip, lastTrip sql.NullString
		var opKm sql.NullFloat64
		if err := rows.Scan(&v.VehicleNo, &v.DailyTrips, &firstTrip, &lastTrip, &opKm); err == nil {
			if firstTrip.Valid {
				v.FirstTripTime = firstTrip.String
			}
			if lastTrip.Valid {
				v.LastTripTime = lastTrip.String
			}
			if opKm.Valid {
				v.OperatingKm = opKm.Float64
			}
			totalTrips += v.DailyTrips
			results = append(results, v)
		}
	}

	avgTrips := 0.0
	if len(results) > 0 {
		avgTrips = float64(totalTrips) / float64(len(results))
	}

	lowCount := 0
	highCount := 0
	for i := range results {
		if float64(results[i].DailyTrips) < avgTrips*0.5 {
			results[i].Level = "low"
			lowCount++
		} else if float64(results[i].DailyTrips) > avgTrips*1.5 {
			results[i].Level = "high"
			highCount++
		} else {
			results[i].Level = "normal"
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].DailyTrips > results[j].DailyTrips
	})

	return results, lowCount, highCount, avgTrips, nil
}

func (s *VehicleService) GetVehicleGantt(vehicleNo, date string) ([]VehicleGanttItem, error) {
	rows, err := db.DB.Query(`
		SELECT vehicle_no, line_no, trip_no, direction,
			actual_departure_time::text as start_time,
			LEAD(actual_departure_time) OVER (PARTITION BY vehicle_no ORDER BY actual_departure_time)::text as end_time,
			trip_date::text
		FROM trips
		WHERE vehicle_no = $1 AND trip_date = $2
		ORDER BY actual_departure_time
	`, vehicleNo, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []VehicleGanttItem
	for rows.Next() {
		var item VehicleGanttItem
		var endTime sql.NullString
		if err := rows.Scan(&item.VehicleNo, &item.LineNo, &item.TripNo, &item.Direction, &item.StartTime, &endTime, &item.TripDate); err == nil {
			if endTime.Valid {
				item.EndTime = endTime.String
			} else {
				var h, m int
				fmt.Sscanf(item.StartTime, "%d:%d", &h, &m)
				m += 45
				if m >= 60 {
					h += m / 60
					m = m % 60
				}
				if h > 23 {
					h = 23
				}
				item.EndTime = fmt.Sprintf("%02d:%02d", h, m)
			}
			results = append(results, item)
		}
	}

	return results, nil
}
