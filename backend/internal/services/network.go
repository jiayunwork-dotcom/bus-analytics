package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"math"
)

type NetworkService struct{}

func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

func (s *NetworkService) GetNetworkMetrics(cityArea, uniqueLineKm float64) (*models.NetworkMetrics, error) {
	metrics := &models.NetworkMetrics{
		CityArea:     cityArea,
		UniqueLineKm: uniqueLineKm,
	}

	db.DB.QueryRow(`SELECT COUNT(DISTINCT station_name) FROM station_flows`).Scan(&metrics.StationCount)

	if cityArea > 0 {
		metrics.StationCoverage = float64(metrics.StationCount) / cityArea * 100
	}

	var totalLineKm sql.NullFloat64
	db.DB.QueryRow(`SELECT SUM(total_km) FROM routes`).Scan(&totalLineKm)
	if totalLineKm.Valid {
		metrics.TotalLineKm = totalLineKm.Float64
	}
	if uniqueLineKm > 0 {
		metrics.LineDuplication = metrics.TotalLineKm / uniqueLineKm
	}

	var nonStraightSum sql.NullFloat64
	var routeCount sql.NullInt64
	db.DB.QueryRow(`
		SELECT SUM(total_km / straight_line_distance_km), COUNT(*)
		FROM routes WHERE straight_line_distance_km IS NOT NULL AND straight_line_distance_km > 0
	`).Scan(&nonStraightSum, &routeCount)
	if nonStraightSum.Valid && routeCount.Valid && routeCount.Int64 > 0 {
		metrics.NonStraightLine = nonStraightSum.Float64 / float64(routeCount.Int64)
	} else {
		metrics.NonStraightLine = 1.4
	}

	metrics.StationCoverage = math.Round(metrics.StationCoverage*100) / 100
	metrics.LineDuplication = math.Round(metrics.LineDuplication*100) / 100
	metrics.NonStraightLine = math.Round(metrics.NonStraightLine*100) / 100

	return metrics, nil
}

type ComparisonService struct{}

func NewComparisonService() *ComparisonService {
	return &ComparisonService{}
}

func (s *ComparisonService) CompareLines(lineNos []string) ([]models.LineEfficiency, error) {
	metricsSvc := NewMetricsService()
	all, err := metricsSvc.GetAllLineEfficiencies()
	if err != nil {
		return nil, err
	}

	lineSet := make(map[string]bool)
	for _, ln := range lineNos {
		lineSet[ln] = true
	}

	var result []models.LineEfficiency
	for _, eff := range all {
		if lineSet[eff.LineNo] {
			result = append(result, eff)
		}
	}
	return result, nil
}

func (s *ComparisonService) GetLineHistoricalTrend(lineNo string, startDate, endDate string, granularity string) ([]map[string]interface{}, error) {
	metricsSvc := NewMetricsService()
	daily, err := metricsSvc.GetLineDailyTrend(lineNo, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if granularity == "day" {
		result := make([]map[string]interface{}, 0, len(daily))
		for _, d := range daily {
			result = append(result, map[string]interface{}{
				"date":                 d.LineName,
				"passenger_intensity":  d.PassengerIntensity,
				"peak_load_factor":     d.PeakLoadFactor,
				"off_peak_load_factor": d.OffPeakLoadFactor,
				"operating_speed":      d.OperatingSpeed,
				"on_time_rate":         d.OnTimeRate,
			})
		}
		return result, nil
	}

	weeklyGroups := make(map[string][]models.LineEfficiency)
	for _, d := range daily {
		datePart := extractDateFromName(d.LineName)
		key := getWeekKey(datePart)
		if granularity == "month" {
			key = getMonthKey(datePart)
		}
		weeklyGroups[key] = append(weeklyGroups[key], d)
	}

	result := make([]map[string]interface{}, 0)
	for key, group := range weeklyGroups {
		agg := map[string]interface{}{
			"period": key,
		}
		var pi, pl, opl, os, otr float64
		for _, g := range group {
			pi += g.PassengerIntensity
			pl += g.PeakLoadFactor
			opl += g.OffPeakLoadFactor
			os += g.OperatingSpeed
			otr += g.OnTimeRate
		}
		n := float64(len(group))
		if n > 0 {
			agg["passenger_intensity"] = math.Round(pi/n*100) / 100
			agg["peak_load_factor"] = math.Round(pl/n*100) / 100
			agg["off_peak_load_factor"] = math.Round(opl/n*100) / 100
			agg["operating_speed"] = math.Round(os/n*100) / 100
			agg["on_time_rate"] = math.Round(otr/n*100) / 100
		}
		result = append(result, agg)
	}
	return result, nil
}

func extractDateFromName(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] == '(' {
			for j := i + 1; j < len(name); j++ {
				if name[j] == ')' {
					return name[i+1 : j]
				}
			}
		}
	}
	return name
}

func getWeekKey(dateStr string) string {
	if len(dateStr) >= 10 {
		return dateStr[:7] + "-W" + getWeekOfMonth(dateStr)
	}
	return dateStr
}

func getWeekOfMonth(dateStr string) string {
	if len(dateStr) >= 10 {
		day := 0
		fmtScan2(dateStr[8:10], &day)
		w := (day-1)/7 + 1
		return itoa2(w)
	}
	return "1"
}

func fmtScan2(s string, d *int) {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			*d = *d*10 + int(c-'0')
		}
	}
}

func itoa2(i int) string {
	return string(rune('0' + i))
}

func getMonthKey(dateStr string) string {
	if len(dateStr) >= 7 {
		return dateStr[:7]
	}
	return dateStr
}
