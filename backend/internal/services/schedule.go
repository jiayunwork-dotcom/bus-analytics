package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"fmt"
	"math"
)

type ScheduleOptimizationService struct{}

func NewScheduleOptimizationService() *ScheduleOptimizationService {
	return &ScheduleOptimizationService{}
}

type TimePeriod struct {
	Name      string
	StartHour int
	EndHour   int
	IsPeak    bool
}

var DefaultPeriods = []TimePeriod{
	{Name: "早高峰", StartHour: 7, EndHour: 9, IsPeak: true},
	{Name: "早平峰", StartHour: 9, EndHour: 12, IsPeak: false},
	{Name: "午间", StartHour: 12, EndHour: 16, IsPeak: false},
	{Name: "晚高峰", StartHour: 16, EndHour: 19, IsPeak: true},
	{Name: "晚平峰", StartHour: 19, EndHour: 21, IsPeak: false},
	{Name: "末班", StartHour: 21, EndHour: 23, IsPeak: false},
}

const (
	MinInterval = 3
	MaxInterval = 30
)

func (s *ScheduleOptimizationService) OptimizeSchedule(lineNo string, totalVehicles int, date string) (*models.ScheduleResult, error) {
	periodData := make([]*periodInfo, 0, len(DefaultPeriods))
	for _, p := range DefaultPeriods {
		passengers, trips := s.getPeriodData(lineNo, date, p.StartHour, p.EndHour)
		hours := float64(p.EndHour - p.StartHour)
		currentInterval := 0
		if trips > 0 && hours > 0 {
			currentInterval = int(hours * 60 / float64(trips))
		}
		if currentInterval < MinInterval {
			currentInterval = MinInterval
		}
		if currentInterval > MaxInterval {
			currentInterval = MaxInterval
		}

		periodData = append(periodData, &periodInfo{
			Period:          p,
			Passengers:      passengers,
			CurrentTrips:    trips,
			CurrentInterval: currentInterval,
			Hours:           hours,
		})
	}

	s.allocateTrips(periodData, totalVehicles)

	result := &models.ScheduleResult{
		Optimizations: make([]models.ScheduleOptimization, 0, len(periodData)),
		TotalTrips:    0,
	}

	totalCurrentWait := 0.0
	totalOptimizedWait := 0.0

	for _, pd := range periodData {
		recInterval := MinInterval
		if pd.RecommendedTrips > 0 && pd.Hours > 0 {
			recInterval = int(math.Round(pd.Hours * 60 / float64(pd.RecommendedTrips)))
		}
		if recInterval < MinInterval {
			recInterval = MinInterval
		}
		if recInterval > MaxInterval {
			recInterval = MaxInterval
		}

		result.Optimizations = append(result.Optimizations, models.ScheduleOptimization{
			PeriodName:          pd.Period.Name,
			TimeRange:           fmt.Sprintf("%02d:00-%02d:00", pd.Period.StartHour, pd.Period.EndHour),
			CurrentInterval:     pd.CurrentInterval,
			RecommendedInterval: recInterval,
			RecommendedTrips:    pd.RecommendedTrips,
			IsPeak:              pd.Period.IsPeak,
		})

		result.TotalTrips += pd.RecommendedTrips
		totalCurrentWait += float64(pd.CurrentInterval) / 2 * pd.Hours
		totalOptimizedWait += float64(recInterval) / 2 * pd.Hours
	}

	if totalCurrentWait > 0 {
		result.CurrentAvgWaitTime = math.Round(totalCurrentWait/6*100) / 100
		result.OptimizedAvgWaitTime = math.Round(totalOptimizedWait/6*100) / 100
		result.ImprovementPercent = math.Round((totalCurrentWait-totalOptimizedWait)/totalCurrentWait*10000) / 100
	}

	return result, nil
}

type periodInfo struct {
	Period            TimePeriod
	Passengers        int
	CurrentTrips      int
	CurrentInterval   int
	RecommendedTrips  int
	Hours             float64
	DemandWeight      float64
}

func (s *ScheduleOptimizationService) allocateTrips(pds []*periodInfo, totalVehicles int) {
	totalWeight := 0.0
	for _, pd := range pds {
		pd.DemandWeight = float64(pd.Passengers+1) / pd.Hours
		if pd.Period.IsPeak {
			pd.DemandWeight *= 2
		}
		totalWeight += pd.DemandWeight
	}

	remainingTrips := totalVehicles * 6
	if remainingTrips <= 0 {
		remainingTrips = 30
	}

	peakMinRatio := 0.4
	peakTrips := int(math.Ceil(float64(remainingTrips) * peakMinRatio))
	offPeakTrips := remainingTrips - peakTrips

	peakWeight := 0.0
	offPeakWeight := 0.0
	for _, pd := range pds {
		if pd.Period.IsPeak {
			peakWeight += pd.DemandWeight
		} else {
			offPeakWeight += pd.DemandWeight
		}
	}

	for _, pd := range pds {
		maxTrips := int(pd.Hours * 60 / MinInterval)
		minTrips := int(math.Ceil(pd.Hours * 60 / MaxInterval))
		if minTrips < 1 {
			minTrips = 1
		}

		weight := pd.DemandWeight
		available := offPeakTrips
		if pd.Period.IsPeak {
			available = peakTrips
			if peakWeight > 0 {
				weight = pd.DemandWeight / peakWeight
			}
		} else {
			if offPeakWeight > 0 {
				weight = pd.DemandWeight / offPeakWeight
			}
		}

		allocated := int(math.Round(float64(available) * weight))
		if allocated < minTrips {
			allocated = minTrips
		}
		if allocated > maxTrips {
			allocated = maxTrips
		}

		pd.RecommendedTrips = allocated
	}

	s.balancePeakOffPeak(pds)
}

func (s *ScheduleOptimizationService) balancePeakOffPeak(pds []*periodInfo) {
	peakIntervals := make([]int, 0)
	offPeakIntervals := make([]int, 0)

	for _, pd := range pds {
		interval := MinInterval
		if pd.RecommendedTrips > 0 && pd.Hours > 0 {
			interval = int(pd.Hours * 60 / float64(pd.RecommendedTrips))
		}
		if pd.Period.IsPeak {
			peakIntervals = append(peakIntervals, interval)
		} else {
			offPeakIntervals = append(offPeakIntervals, interval)
		}
	}

	for _, pd := range pds {
		if !pd.Period.IsPeak {
			continue
		}
		peakInterval := MinInterval
		if pd.RecommendedTrips > 0 && pd.Hours > 0 {
			peakInterval = int(pd.Hours * 60 / float64(pd.RecommendedTrips))
		}
		for _, oi := range offPeakIntervals {
			for peakInterval > oi && pd.RecommendedTrips > 1 {
				pd.RecommendedTrips--
				peakInterval = int(pd.Hours * 60 / float64(pd.RecommendedTrips))
			}
		}
	}
}

func (s *ScheduleOptimizationService) getPeriodData(lineNo, date string, startHour, endHour int) (int, int) {
	var passengers sql.NullInt64
	db.DB.QueryRow(`
		SELECT SUM(s.board_count)
		FROM station_flows s
		INNER JOIN trips t ON s.line_no = t.line_no AND s.trip_no = t.trip_no AND s.flow_date = t.trip_date
		WHERE s.line_no = $1 AND s.flow_date = $2
			AND EXTRACT(HOUR FROM t.actual_departure_time) >= $3
			AND EXTRACT(HOUR FROM t.actual_departure_time) < $4
	`, lineNo, date, startHour, endHour).Scan(&passengers)

	var trips sql.NullInt64
	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT trip_no)
		FROM trips
		WHERE line_no = $1 AND trip_date = $2
			AND EXTRACT(HOUR FROM actual_departure_time) >= $3
			AND EXTRACT(HOUR FROM actual_departure_time) < $4
	`, lineNo, date, startHour, endHour).Scan(&trips)

	p := 0
	t := 0
	if passengers.Valid {
		p = int(passengers.Int64)
	}
	if trips.Valid {
		t = int(trips.Int64)
	}
	return p, t
}
