package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"fmt"
	"math"
	"sort"
)

const RatedCapacity = 80

func parseHour(timeStr string) int {
	hour := 0
	for i := 0; i < len(timeStr) && i < 2; i++ {
		c := timeStr[i]
		if c >= '0' && c <= '9' {
			hour = hour*10 + int(c-'0')
		} else {
			break
		}
	}
	if hour < 0 || hour > 23 {
		return 0
	}
	return hour
}

type MetricsService struct{}

func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

type dailyMetric struct {
	Date        string
	Passengers  int
	PeakLoad    float64
	OffPeakLoad float64
	OperatingSpeed float64
	OnTimeRate  float64
}

func (s *MetricsService) GetAllLineEfficiencies() ([]models.LineEfficiency, error) {
	rows, err := db.DB.Query(`
		SELECT DISTINCT r.line_no, r.line_name, r.total_km
		FROM routes r
		INNER JOIN trips t ON r.line_no = t.line_no
		ORDER BY r.line_no
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type lineInfo struct {
		lineNo   string
		lineName string
		totalKm  float64
	}
	var lines []lineInfo
	for rows.Next() {
		var l lineInfo
		if err := rows.Scan(&l.lineNo, &l.lineName, &l.totalKm); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}

	results := make([]models.LineEfficiency, 0, len(lines))
	avgs := map[string]float64{
		"passenger_intensity": 0,
		"peak_load":           0,
		"speed":               0,
		"on_time":             0,
	}
	count := 0

	for _, l := range lines {
		eff, err := s.calcLineEfficiency(l.lineNo, l.lineName, l.totalKm)
		if err != nil {
			continue
		}
		results = append(results, *eff)
		avgs["passenger_intensity"] += eff.PassengerIntensity
		avgs["peak_load"] += eff.PeakLoadFactor
		avgs["speed"] += eff.OperatingSpeed
		avgs["on_time"] += eff.OnTimeRate
		count++
	}

	if count > 0 {
		for k := range avgs {
			avgs[k] /= float64(count)
		}
	}

	for i := range results {
		results[i].Level = s.calcLevel(results[i], avgs)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].LineNo < results[j].LineNo
	})

	return results, nil
}

func (s *MetricsService) calcLineEfficiency(lineNo, lineName string, totalKm float64) (*models.LineEfficiency, error) {
	dailyPassengers := s.getDailyPassengers(lineNo)
	daysCount := len(dailyPassengers)
	if daysCount == 0 {
		daysCount = 1
	}

	totalDailyPassengers := 0.0
	for _, p := range dailyPassengers {
		totalDailyPassengers += float64(p)
	}
	avgDailyPassengers := totalDailyPassengers / float64(daysCount)
	passengerIntensity := 0.0
	if totalKm > 0 {
		passengerIntensity = avgDailyPassengers / totalKm
	}

	peakLoad, offPeakLoad := s.calcLoadFactors(lineNo)
	operatingSpeed := s.calcOperatingSpeed(lineNo)
	onTimeRate := s.calcOnTimeRate(lineNo)

	return &models.LineEfficiency{
		LineNo:             lineNo,
		LineName:           lineName,
		PassengerIntensity: passengerIntensity,
		PeakLoadFactor:     peakLoad,
		OffPeakLoadFactor:  offPeakLoad,
		OperatingSpeed:     operatingSpeed,
		OnTimeRate:         onTimeRate,
	}, nil
}

func (s *MetricsService) getDailyPassengers(lineNo string) map[string]int {
	result := make(map[string]int)
	rows, err := db.DB.Query(`
		SELECT flow_date::text, SUM(board_count)
		FROM station_flows
		WHERE line_no = $1
		GROUP BY flow_date
	`, lineNo)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var d string
		var c int
		if err := rows.Scan(&d, &c); err == nil {
			result[d] = c
		}
	}
	return result
}

func (s *MetricsService) calcLoadFactors(lineNo string) (float64, float64) {
	peakSum := 0.0
	peakCount := 0
	offPeakSum := 0.0
	offPeakCount := 0

	rows, err := db.DB.Query(`
		SELECT DISTINCT ON (t.trip_no, t.trip_date)
			t.actual_departure_time::text, t.trip_no, t.trip_date::text
		FROM trips t
		WHERE t.line_no = $1
		ORDER BY t.trip_date, t.trip_no, t.actual_departure_time
	`, lineNo)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	for rows.Next() {
		var depTime, tripNo, tripDate string
		if err := rows.Scan(&depTime, &tripNo, &tripDate); err != nil {
			continue
		}

		maxPax := s.getMaxOnBoard(lineNo, tripNo, tripDate, 0)
		load := float64(maxPax) / RatedCapacity
		if load > 1.0 {
			load = 1.0
		}

		hour := parseHour(depTime)
		if (hour >= 7 && hour < 9) || (hour >= 17 && hour < 19) {
			peakSum += load
			peakCount++
		} else {
			offPeakSum += load
			offPeakCount++
		}
	}

	peakLoad := 0.0
	if peakCount > 0 {
		peakLoad = peakSum / float64(peakCount)
	}
	offPeakLoad := 0.0
	if offPeakCount > 0 {
		offPeakLoad = offPeakSum / float64(offPeakCount)
	}
	return peakLoad, offPeakLoad
}

func (s *MetricsService) getMaxOnBoard(lineNo, tripNo, flowDate string, direction int) int {
	rows, err := db.DB.Query(`
		SELECT station_seq, board_count, alight_count
		FROM station_flows
		WHERE line_no = $1 AND trip_no = $2 AND flow_date = $3::date
		ORDER BY station_seq
	`, lineNo, tripNo, flowDate)
	if err != nil {
		return 0
	}
	defer rows.Close()

	onBoard := 0
	maxOnBoard := 0
	for rows.Next() {
		var seq, board, alight int
		if err := rows.Scan(&seq, &board, &alight); err == nil {
			effectiveAlight := alight
			if effectiveAlight > onBoard {
				effectiveAlight = onBoard
			}
			onBoard = onBoard - effectiveAlight + board
			if onBoard < 0 {
				onBoard = 0
			}
			if onBoard > maxOnBoard {
				maxOnBoard = onBoard
			}
		}
	}
	return maxOnBoard
}

func (s *MetricsService) calcOperatingSpeed(lineNo string) float64 {
	var totalKm sql.NullFloat64
	db.DB.QueryRow(`
		SELECT SUM(vm.operating_km)
		FROM vehicle_mileages vm
		INNER JOIN trips t ON vm.vehicle_no = t.vehicle_no AND vm.mileage_date = t.trip_date
		WHERE t.line_no = $1
	`, lineNo).Scan(&totalKm)

	var firstTime, lastTime sql.NullString
	db.DB.QueryRow(`
		SELECT MIN(actual_departure_time)::text, MAX(actual_departure_time)::text
		FROM trips
		WHERE line_no = $1
	`, lineNo).Scan(&firstTime, &lastTime)

	if !totalKm.Valid || !firstTime.Valid || !lastTime.Valid {
		return 0
	}

	ft, err1 := ParseTime(firstTime.String)
	lt, err2 := ParseTime(lastTime.String)
	if err1 != nil || err2 != nil {
		return 0
	}

	hours := lt.Sub(ft).Hours()
	if hours <= 0 {
		return 0
	}

	return totalKm.Float64 / hours
}

func (s *MetricsService) calcOnTimeRate(lineNo string) float64 {
	var total, onTime sql.NullInt64
	db.DB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (
				WHERE ABS(
					EXTRACT(EPOCH FROM (actual_departure_time - planned_departure_time))
				) <= 180
			) as on_time
		FROM trips
		WHERE line_no = $1
	`, lineNo).Scan(&total, &onTime)

	if !total.Valid || total.Int64 == 0 {
		return 0
	}
	return float64(onTime.Int64) / float64(total.Int64) * 100
}

func (s *MetricsService) calcLevel(eff models.LineEfficiency, avgs map[string]float64) string {
	goodCount := 0
	badCount := 0

	if eff.PassengerIntensity > avgs["passenger_intensity"]*1.2 {
		goodCount++
	} else if eff.PassengerIntensity < avgs["passenger_intensity"]*0.7 {
		badCount++
	}
	if eff.PeakLoadFactor > avgs["peak_load"]*1.2 {
		goodCount++
	} else if eff.PeakLoadFactor < avgs["peak_load"]*0.7 {
		badCount++
	}
	if eff.OperatingSpeed > avgs["speed"]*1.2 {
		goodCount++
	} else if eff.OperatingSpeed < avgs["speed"]*0.7 {
		badCount++
	}
	if eff.OnTimeRate > avgs["on_time"]*1.2 {
		goodCount++
	} else if eff.OnTimeRate < avgs["on_time"]*0.7 {
		badCount++
	}

	if goodCount >= 2 {
		return "green"
	} else if badCount >= 2 {
		return "red"
	}
	return "yellow"
}

func (s *MetricsService) GetLineDailyTrend(lineNo string, startDate, endDate string) ([]models.LineEfficiency, error) {
	var lineName string
	var totalKm float64
	db.DB.QueryRow(`SELECT line_name, total_km FROM routes WHERE line_no = $1`, lineNo).Scan(&lineName, &totalKm)

	query := `
		SELECT DISTINCT trip_date::text
		FROM trips
		WHERE line_no = $1
	`
	args := []interface{}{lineNo}
	argIdx := 2
	if startDate != "" {
		query += fmt.Sprintf(" AND trip_date >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND trip_date <= $%d", argIdx)
		args = append(args, endDate)
		argIdx++
	}
	query += " ORDER BY trip_date"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err == nil {
			dates = append(dates, d)
		}
	}

	var results []models.LineEfficiency
	for _, d := range dates {
		eff := models.LineEfficiency{
			LineNo:   lineNo,
			LineName: lineName + "(" + d + ")",
		}

		var dailyPassengers sql.NullInt64
		db.DB.QueryRow(`SELECT SUM(board_count) FROM station_flows WHERE line_no = $1 AND flow_date = $2::date`, lineNo, d).Scan(&dailyPassengers)
		if totalKm > 0 && dailyPassengers.Valid {
			eff.PassengerIntensity = float64(dailyPassengers.Int64) / totalKm
		}

		peakLoad, offPeakLoad := s.calcDailyLoadFactor(lineNo, d)
		eff.PeakLoadFactor = peakLoad
		eff.OffPeakLoadFactor = offPeakLoad

		eff.OperatingSpeed = s.calcDailyOperatingSpeed(lineNo, d)
		eff.OnTimeRate = s.calcDailyOnTimeRate(lineNo, d)

		results = append(results, eff)
	}

	return results, nil
}

func (s *MetricsService) calcDailyLoadFactor(lineNo, date string) (float64, float64) {
	rows, err := db.DB.Query(`
		SELECT DISTINCT ON (t.trip_no)
			t.actual_departure_time::text, t.trip_no
		FROM trips t
		WHERE t.line_no = $1 AND t.trip_date = $2::date
		ORDER BY t.trip_no, t.actual_departure_time
	`, lineNo, date)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	peakSum, offPeakSum := 0.0, 0.0
	peakCnt, offPeakCnt := 0, 0
	for rows.Next() {
		var depTime, tripNo string
		if err := rows.Scan(&depTime, &tripNo); err != nil {
			continue
		}
		maxPax := s.getMaxOnBoard(lineNo, tripNo, date, 0)
		load := float64(maxPax) / RatedCapacity
		if load > 1.0 {
			load = 1.0
		}
		hour := parseHour(depTime)
		if (hour >= 7 && hour < 9) || (hour >= 17 && hour < 19) {
			peakSum += load
			peakCnt++
		} else {
			offPeakSum += load
			offPeakCnt++
		}
	}
	peak := 0.0
	if peakCnt > 0 {
		peak = peakSum / float64(peakCnt)
	}
	offPeak := 0.0
	if offPeakCnt > 0 {
		offPeak = offPeakSum / float64(offPeakCnt)
	}
	return peak, offPeak
}

func (s *MetricsService) calcDailyOperatingSpeed(lineNo, date string) float64 {
	var totalKm sql.NullFloat64
	db.DB.QueryRow(`
		SELECT SUM(vm.operating_km)
		FROM vehicle_mileages vm
		INNER JOIN trips t ON vm.vehicle_no = t.vehicle_no AND vm.mileage_date = t.trip_date
		WHERE t.line_no = $1 AND t.trip_date = $2::date
	`, lineNo, date).Scan(&totalKm)

	var firstTime, lastTime sql.NullString
	db.DB.QueryRow(`
		SELECT MIN(actual_departure_time)::text, MAX(actual_departure_time)::text
		FROM trips WHERE line_no = $1 AND trip_date = $2::date
	`, lineNo, date).Scan(&firstTime, &lastTime)

	if !totalKm.Valid || !firstTime.Valid || !lastTime.Valid {
		return 0
	}
	ft, err1 := ParseTime(firstTime.String)
	lt, err2 := ParseTime(lastTime.String)
	if err1 != nil || err2 != nil {
		return 0
	}
	hours := lt.Sub(ft).Hours()
	if hours <= 0 {
		return 0
	}
	return totalKm.Float64 / hours
}

func (s *MetricsService) calcDailyOnTimeRate(lineNo, date string) float64 {
	var total, onTime sql.NullInt64
	db.DB.QueryRow(`
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE ABS(EXTRACT(EPOCH FROM (actual_departure_time - planned_departure_time))) <= 180)
		FROM trips WHERE line_no = $1 AND trip_date = $2::date
	`, lineNo, date).Scan(&total, &onTime)
	if !total.Valid || total.Int64 == 0 {
		return 0
	}
	return math.Round(float64(onTime.Int64)/float64(total.Int64)*10000) / 100
}
