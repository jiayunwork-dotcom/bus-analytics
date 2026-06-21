package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
)

type SimulationService struct {
	scheduleSvc *ScheduleOptimizationService
	metricsSvc  *MetricsService
	flowSvc     *FlowAnalysisService
}

func NewSimulationService() *SimulationService {
	return &SimulationService{
		scheduleSvc: NewScheduleOptimizationService(),
		metricsSvc:  NewMetricsService(),
		flowSvc:     NewFlowAnalysisService(),
	}
}

type stationFlow struct {
	StationSeq  int
	StationName string
	BoardCount  int
	AlightCount int
}

func (s *SimulationService) RunSimulation(params *models.SimParams) (*models.SimResult, error) {
	lineInfo := s.getLineInfo(params.LineNo)
	if lineInfo == nil {
		return nil, fmt.Errorf("线路不存在: %s", params.LineNo)
	}

	date := params.Date
	if date == "" {
		date = s.getLatestDate(params.LineNo)
	}

	availableVehicles := s.getAvailableVehicles()
	origStationCount := lineInfo.StationCount
	newStationCount := origStationCount + params.StationDelta
	if newStationCount < 2 {
		newStationCount = 2
	}

	origIntervals := s.getCurrentIntervals(params.LineNo, date)
	origPeakInterval := origIntervals.peak
	origOffPeakInterval := origIntervals.offPeak

	newPeakInterval := params.PeakInterval
	if newPeakInterval == 0 {
		newPeakInterval = origPeakInterval
	}
	newOffPeakInterval := params.OffPeakInterval
	if newOffPeakInterval == 0 {
		newOffPeakInterval = origOffPeakInterval
	}

	newPeakInterval = clampInterval(newPeakInterval)
	newOffPeakInterval = clampInterval(newOffPeakInterval)
	if newPeakInterval > newOffPeakInterval {
		newPeakInterval = newOffPeakInterval
	}

	origTotalTrips := s.calcDailyTrips(params.LineNo, date, origPeakInterval, origOffPeakInterval)
	newTotalTrips := s.calcDailyTripsWithInterval(params.LineNo, date, newPeakInterval, newOffPeakInterval)

	peakHours := s.getPeakHours()
	origPeakTrips := int(math.Ceil(peakHours * 60 / float64(origPeakInterval)))
	newPeakTrips := int(math.Ceil(peakHours * 60 / float64(newPeakInterval)))

	capacityWarning := newPeakTrips > availableVehicles

	peakCapacityChange := 0.0
	if origPeakTrips > 0 {
		peakCapacityChange = math.Round(float64(newPeakTrips-origPeakTrips)/float64(origPeakTrips)*10000) / 100
	}

	origDailyPassengers := s.getDailyPassengers(params.LineNo, date)
	origPeakLoad, _ := s.metricsSvc.calcDailyLoadFactor(params.LineNo, date)
	origOperatingSpeed := s.calcDailyOperatingSpeedSafe(params.LineNo, date)
	origPassengerIntensity := 0.0
	if lineInfo.TotalKm > 0 {
		origPassengerIntensity = float64(origDailyPassengers) / lineInfo.TotalKm
	}

	origKPI := models.SimKPI{
		DailyTrips:         origTotalTrips,
		PeakLoadFactor:     origPeakLoad,
		OperatingSpeed:     origOperatingSpeed,
		PassengerIntensity: origPassengerIntensity,
	}

	stationDelta := params.StationDelta
	flows := s.getAggregatedStationFlows(params.LineNo, date)
	sharedStations := s.getSharedStationNames(params.LineNo)
	adjustedFlows := s.adjustStationFlows(flows, stationDelta, sharedStations)

	newDailyPassengers := origDailyPassengers
	if stationDelta < 0 {
		removeCount := -stationDelta
		removeIndexes := s.getRemoveIndexes(flows, removeCount, sharedStations)
		for _, idx := range removeIndexes {
			if idx >= 0 && idx < len(flows) {
				newDailyPassengers -= int(float64(flows[idx].BoardCount) * 0.3)
			}
		}
		if newDailyPassengers < 0 {
			newDailyPassengers = 0
		}
	}

	newTotalKm := lineInfo.TotalKm
	if origStationCount > 0 {
		newTotalKm = lineInfo.TotalKm * float64(newStationCount) / float64(origStationCount)
	}

	newPeakLoad := s.calcNewPeakLoad(adjustedFlows, newTotalTrips, origPeakLoad, newPeakInterval, origPeakInterval)

	newOperatingSpeed := origOperatingSpeed
	if origStationCount > 0 && origOperatingSpeed > 0 {
		newOperatingSpeed = origOperatingSpeed * float64(origStationCount) / float64(newStationCount)
	}

	newPassengerIntensity := 0.0
	if newTotalKm > 0 {
		newPassengerIntensity = float64(newDailyPassengers) / newTotalKm
	}

	newKPI := models.SimKPI{
		DailyTrips:         newTotalTrips,
		PeakLoadFactor:     newPeakLoad,
		OperatingSpeed:     newOperatingSpeed,
		PassengerIntensity: newPassengerIntensity,
	}

	adjacentImpacts := s.calcAdjacentImpacts(params.LineNo, stationDelta, flows, sharedStations, date)

	removalTrend := s.calcRemovalTrend(params.LineNo, date, flows, sharedStations, origStationCount, origPeakLoad, origPeakInterval)

	return &models.SimResult{
		LineNo:              params.LineNo,
		LineName:            lineInfo.LineName,
		OrigStationCount:    origStationCount,
		NewStationCount:     newStationCount,
		OrigPeakInterval:    origPeakInterval,
		NewPeakInterval:     newPeakInterval,
		OrigOffPeakInterval: origOffPeakInterval,
		NewOffPeakInterval:  newOffPeakInterval,
		OrigTotalTrips:      origTotalTrips,
		NewTotalTrips:       newTotalTrips,
		TripsDelta:          newTotalTrips - origTotalTrips,
		PeakCapacityChange:  peakCapacityChange,
		CapacityWarning:     capacityWarning,
		AvailableVehicles:   availableVehicles,
		OrigKPI:             origKPI,
		NewKPI:              newKPI,
		AdjacentImpacts:     adjacentImpacts,
		RemovalTrend:        removalTrend,
		OrigPeakTrips:       origPeakTrips,
		NewPeakTrips:        newPeakTrips,
	}, nil
}

type routeInfo struct {
	LineNo       string
	LineName     string
	TotalKm      float64
	StationCount int
}

func (s *SimulationService) getLineInfo(lineNo string) *routeInfo {
	var r routeInfo
	err := db.DB.QueryRow(`
		SELECT line_no, line_name, total_km, station_count
		FROM routes WHERE line_no = $1
	`, lineNo).Scan(&r.LineNo, &r.LineName, &r.TotalKm, &r.StationCount)
	if err != nil {
		return nil
	}
	return &r
}

func (s *SimulationService) getLatestDate(lineNo string) string {
	var d sql.NullString
	db.DB.QueryRow(`SELECT MAX(trip_date)::text FROM trips WHERE line_no = $1`, lineNo).Scan(&d)
	if d.Valid {
		return d.String
	}
	return ""
}

func (s *SimulationService) getAvailableVehicles() int {
	var cnt sql.NullInt64
	db.DB.QueryRow(`SELECT COUNT(DISTINCT vehicle_no) FROM trips`).Scan(&cnt)
	if cnt.Valid && cnt.Int64 > 0 {
		return int(cnt.Int64)
	}
	return 20
}

type intervalPair struct {
	peak    int
	offPeak int
}

func (s *SimulationService) getCurrentIntervals(lineNo, date string) intervalPair {
	peakTrips := 0
	offPeakTrips := 0
	peakHours := 0.0
	offPeakHours := 0.0

	for _, p := range DefaultPeriods {
		hours := float64(p.EndHour - p.StartHour)
		_, trips := s.scheduleSvc.getPeriodData(lineNo, date, p.StartHour, p.EndHour)
		if p.IsPeak {
			peakTrips += trips
			peakHours += hours
		} else {
			offPeakTrips += trips
			offPeakHours += hours
		}
	}

	peakInterval := MinInterval
	if peakTrips > 0 && peakHours > 0 {
		peakInterval = int(math.Round(peakHours * 60 / float64(peakTrips)))
	}
	peakInterval = clampInterval(peakInterval)

	offPeakInterval := MinInterval
	if offPeakTrips > 0 && offPeakHours > 0 {
		offPeakInterval = int(math.Round(offPeakHours * 60 / float64(offPeakTrips)))
	}
	offPeakInterval = clampInterval(offPeakInterval)

	return intervalPair{peak: peakInterval, offPeak: offPeakInterval}
}

func clampInterval(iv int) int {
	if iv < MinInterval {
		return MinInterval
	}
	if iv > MaxInterval {
		return MaxInterval
	}
	return iv
}

func (s *SimulationService) getPeakHours() float64 {
	hours := 0.0
	for _, p := range DefaultPeriods {
		if p.IsPeak {
			hours += float64(p.EndHour - p.StartHour)
		}
	}
	return hours
}

func (s *SimulationService) getOffPeakHours() float64 {
	hours := 0.0
	for _, p := range DefaultPeriods {
		if !p.IsPeak {
			hours += float64(p.EndHour - p.StartHour)
		}
	}
	return hours
}

func (s *SimulationService) calcDailyTrips(lineNo, date string, peakInterval, offPeakInterval int) int {
	var total sql.NullInt64
	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT trip_no) FROM trips
		WHERE line_no = $1 AND trip_date = $2::date
	`, lineNo, date).Scan(&total)
	if total.Valid && total.Int64 > 0 {
		return int(total.Int64)
	}
	return s.calcDailyTripsWithInterval(lineNo, date, peakInterval, offPeakInterval)
}

func (s *SimulationService) calcDailyTripsWithInterval(lineNo, date string, peakInterval, offPeakInterval int) int {
	peakHours := s.getPeakHours()
	offPeakHours := s.getOffPeakHours()

	peakTrips := 0
	if peakInterval > 0 {
		peakTrips = int(math.Ceil(peakHours * 60 / float64(peakInterval)))
	}
	offPeakTrips := 0
	if offPeakInterval > 0 {
		offPeakTrips = int(math.Ceil(offPeakHours * 60 / float64(offPeakInterval)))
	}

	maxTrips := 0
	var cnt sql.NullInt64
	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT trip_no) FROM trips
		WHERE line_no = $1 AND trip_date = $2::date
	`, lineNo, date).Scan(&cnt)
	if cnt.Valid {
		maxTrips = int(cnt.Int64)
	}

	estimated := peakTrips + offPeakTrips
	if maxTrips > 0 {
		ratio := float64(maxTrips) / float64(estimated+1)
		if ratio > 1.5 {
			ratio = 1.5
		}
		if ratio < 0.5 {
			ratio = 0.5
		}
		return int(math.Round(float64(estimated) * ratio))
	}
	return estimated
}

func (s *SimulationService) getDailyPassengers(lineNo, date string) int {
	var total sql.NullInt64
	db.DB.QueryRow(`
		SELECT SUM(board_count) FROM station_flows
		WHERE line_no = $1 AND flow_date = $2::date
	`, lineNo, date).Scan(&total)
	if total.Valid {
		return int(total.Int64)
	}
	return 0
}

func (s *SimulationService) getAggregatedStationFlows(lineNo, date string) []stationFlow {
	rows, err := db.DB.Query(`
		SELECT station_seq, station_name,
			COALESCE(SUM(board_count), 0) as board,
			COALESCE(SUM(alight_count), 0) as alight
		FROM station_flows
		WHERE line_no = $1 AND flow_date = $2::date
		GROUP BY station_seq, station_name
		ORDER BY station_seq
	`, lineNo, date)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []stationFlow
	tripCount := s.getTripCount(lineNo, date)
	for rows.Next() {
		var f stationFlow
		if err := rows.Scan(&f.StationSeq, &f.StationName, &f.BoardCount, &f.AlightCount); err == nil {
			if tripCount > 0 {
				f.BoardCount = f.BoardCount / tripCount
				f.AlightCount = f.AlightCount / tripCount
			}
			result = append(result, f)
		}
	}
	return result
}

func (s *SimulationService) getTripCount(lineNo, date string) int {
	var cnt sql.NullInt64
	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT trip_no) FROM trips
		WHERE line_no = $1 AND trip_date = $2::date
	`, lineNo, date).Scan(&cnt)
	if cnt.Valid {
		return int(cnt.Int64)
	}
	return 1
}

func (s *SimulationService) calcDailyOperatingSpeedSafe(lineNo, date string) float64 {
	var totalKm sql.NullFloat64
	db.DB.QueryRow(`SELECT total_km FROM routes WHERE line_no = $1`, lineNo).Scan(&totalKm)
	if !totalKm.Valid || totalKm.Float64 <= 0 {
		return 0
	}

	var avgInterval sql.NullFloat64
	db.DB.QueryRow(`
		WITH vehicle_trips AS (
			SELECT vehicle_no, actual_departure_time,
				LAG(actual_departure_time) OVER (PARTITION BY vehicle_no ORDER BY actual_departure_time) as prev_dep
			FROM trips
			WHERE line_no = $1 AND trip_date = $2::date
		)
		SELECT AVG(EXTRACT(EPOCH FROM (actual_departure_time - prev_dep))/3600)
		FROM vehicle_trips
		WHERE prev_dep IS NOT NULL
	`, lineNo, date).Scan(&avgInterval)

	if !avgInterval.Valid || avgInterval.Float64 <= 0 {
		var stationCount sql.NullInt64
		db.DB.QueryRow(`SELECT station_count FROM routes WHERE line_no = $1`, lineNo).Scan(&stationCount)
		if stationCount.Valid && stationCount.Int64 > 0 {
			estimatedTripHours := float64(stationCount.Int64) * 3.0 / 60.0
			if estimatedTripHours > 0 {
				speed := totalKm.Float64 / estimatedTripHours
				return math.Round(speed*100) / 100
			}
		}
		return 0
	}

	estimatedTripHours := avgInterval.Float64 * 0.85
	speed := totalKm.Float64 / estimatedTripHours
	return math.Round(speed*100) / 100
}

func (s *SimulationService) getSharedStationNames(lineNo string) map[string]bool {
	shared := make(map[string]bool)
	rows, err := db.DB.Query(`
		SELECT DISTINCT sf.station_name
		FROM station_flows sf
		WHERE sf.line_no = $1
		AND EXISTS (
			SELECT 1 FROM station_flows sf2
			WHERE sf2.station_name = sf.station_name
			AND sf2.line_no <> sf.line_no
		)
	`, lineNo)
	if err != nil {
		return shared
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			shared[name] = true
		}
	}
	return shared
}

func (s *SimulationService) getRemoveIndexes(flows []stationFlow, removeCount int, sharedStations map[string]bool) []int {
	result := []int{}
	if removeCount <= 0 || len(flows) <= 2 {
		return result
	}
	maxRemove := len(flows) - 2
	if removeCount > maxRemove {
		removeCount = maxRemove
	}

	allIndexes := []int{}
	for i := 1; i < len(flows)-1; i++ {
		allIndexes = append(allIndexes, i)
	}

	sharedIndexes := []int{}
	nonSharedIndexes := []int{}
	for _, idx := range allIndexes {
		if sharedStations[flows[idx].StationName] {
			sharedIndexes = append(sharedIndexes, idx)
		} else {
			nonSharedIndexes = append(nonSharedIndexes, idx)
		}
	}

	for i := len(sharedIndexes) - 1; i >= 0 && len(result) < removeCount; i-- {
		result = append(result, sharedIndexes[i])
	}

	for i := len(nonSharedIndexes) - 1; i >= 0 && len(result) < removeCount; i-- {
		result = append(result, nonSharedIndexes[i])
	}

	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result
}

func (s *SimulationService) adjustStationFlows(flows []stationFlow, delta int, sharedStations map[string]bool) []stationFlow {
	if len(flows) == 0 || delta == 0 {
		return flows
	}

	result := make([]stationFlow, len(flows))
	for i, f := range flows {
		result[i] = stationFlow{
			StationSeq:  f.StationSeq,
			StationName: f.StationName,
			BoardCount:  f.BoardCount,
			AlightCount: f.AlightCount,
		}
	}

	if delta < 0 {
		removeCount := -delta
		indexes := s.getRemoveIndexes(result, removeCount, sharedStations)
		for _, removeIdx := range indexes {
			if removeIdx < 1 || removeIdx >= len(result)-1 {
				continue
			}
			removed := result[removeIdx]
			prevIdx := removeIdx - 1
			nextIdx := removeIdx + 1
			if prevIdx >= 0 {
				result[prevIdx].BoardCount += removed.BoardCount / 2
				result[prevIdx].AlightCount += removed.AlightCount / 2
			}
			if nextIdx < len(result) {
				result[nextIdx].BoardCount += removed.BoardCount / 2
				result[nextIdx].AlightCount += removed.AlightCount / 2
			}
			result = append(result[:removeIdx], result[removeIdx+1:]...)
		}
	} else if delta > 0 {
		lastSeq := 0
		if len(result) > 0 {
			lastSeq = result[len(result)-1].StationSeq
		}
		for i := 0; i < delta; i++ {
			lastSeq++
			result = append(result, stationFlow{
				StationSeq:  lastSeq,
				StationName: fmt.Sprintf("新增站%d", i+1),
				BoardCount:  5,
				AlightCount: 5,
			})
		}
	}

	return result
}

func (s *SimulationService) calcNewPeakLoad(flows []stationFlow, newTrips int, origLoad float64, newPeakInterval, origPeakInterval int) float64 {
	if len(flows) == 0 {
		return origLoad
	}

	onBoard := 0
	maxOnBoard := 0
	for _, f := range flows {
		effectiveAlight := f.AlightCount
		if effectiveAlight > onBoard {
			effectiveAlight = onBoard
		}
		onBoard = onBoard - effectiveAlight + f.BoardCount
		if onBoard < 0 {
			onBoard = 0
		}
		if onBoard > maxOnBoard {
			maxOnBoard = onBoard
		}
	}

	peakRatio := 1.0
	if origPeakInterval > 0 {
		peakRatio = float64(origPeakInterval) / float64(newPeakInterval)
	}
	adjustedMax := float64(maxOnBoard) * peakRatio
	load := adjustedMax / RatedCapacity
	if load > 1.0 {
		load = 1.0
	}
	if load < 0 {
		load = 0
	}
	return math.Round(load*10000) / 10000
}

func (s *SimulationService) calcAdjacentImpacts(targetLineNo string, stationDelta int, origFlows []stationFlow, sharedStations map[string]bool, date string) []models.AdjLineImpact {
	result := []models.AdjLineImpact{}

	if stationDelta >= 0 || len(origFlows) == 0 {
		return result
	}

	removeCount := -stationDelta
	removeIndexes := s.getRemoveIndexes(origFlows, removeCount, sharedStations)

	removedStations := make(map[string]int)
	for _, idx := range removeIndexes {
		if idx >= 0 && idx < len(origFlows) {
			removedStations[origFlows[idx].StationName] = origFlows[idx].BoardCount
		}
	}

	allLines := s.getAllLines()
	for _, line := range allLines {
		if line.LineNo == targetLineNo {
			continue
		}
		sharedWithAdj := s.getSharedStations(line.LineNo, removedStations)
		if len(sharedWithAdj) == 0 {
			continue
		}

		var origPeakLoad float64
		origPeakLoad, _ = s.metricsSvc.calcDailyLoadFactor(line.LineNo, date)

		transferPax := 0
		for _, st := range sharedWithAdj {
			if cnt, ok := removedStations[st]; ok {
				transferPax += int(float64(cnt) * 0.3)
			}
		}

		adjTrips := s.getTripCount(line.LineNo, date)
		additionalLoad := 0.0
		if adjTrips > 0 {
			additionalLoad = float64(transferPax) / float64(adjTrips) / RatedCapacity
		}

		newPeakLoad := origPeakLoad + additionalLoad
		if newPeakLoad > 1.0 {
			newPeakLoad = 1.0
		}
		overloadRisk := newPeakLoad > 0.9

		result = append(result, models.AdjLineImpact{
			LineNo:         line.LineNo,
			LineName:       line.LineName,
			OrigPeakLoad:   origPeakLoad,
			NewPeakLoad:    math.Round(newPeakLoad*10000) / 10000,
			LoadIncrement:  math.Round(additionalLoad*10000) / 10000,
			OverloadRisk:   overloadRisk,
			SharedStations: sharedWithAdj,
		})
	}

	return result
}

type lineSimple struct {
	LineNo   string
	LineName string
}

func (s *SimulationService) getAllLines() []lineSimple {
	rows, err := db.DB.Query(`SELECT line_no, line_name FROM routes ORDER BY line_no`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []lineSimple
	for rows.Next() {
		var l lineSimple
		if err := rows.Scan(&l.LineNo, &l.LineName); err == nil {
			result = append(result, l)
		}
	}
	return result
}

func (s *SimulationService) getSharedStations(lineNo string, stations map[string]int) []string {
	rows, err := db.DB.Query(`
		SELECT DISTINCT station_name FROM station_flows WHERE line_no = $1
	`, lineNo)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			if _, ok := stations[name]; ok {
				result = append(result, name)
			}
		}
	}
	return result
}

func (s *SimulationService) calcRemovalTrend(lineNo, date string, flows []stationFlow, sharedStations map[string]bool, origStationCount int, origPeakLoad float64, origPeakInterval int) []models.TrendPoint {
	result := []models.TrendPoint{}

	result = append(result, models.TrendPoint{
		RemoveCount:    0,
		PeakLoadFactor: origPeakLoad,
	})

	for removeCount := 1; removeCount <= 5; removeCount++ {
		adjustedFlows := s.adjustStationFlows(flows, -removeCount, sharedStations)
		load := s.calcNewPeakLoad(adjustedFlows, 0, origPeakLoad, origPeakInterval, origPeakInterval)

		result = append(result, models.TrendPoint{
			RemoveCount:    removeCount,
			PeakLoadFactor: load,
		})
	}

	return result
}

var lineColors = []string{"#409eff", "#67c23a", "#e6a23c", "#f56c6c", "#909399"}

func (s *SimulationService) RunJointSimulation(params *models.JointSimParams) (*models.JointSimResult, error) {
	if len(params.Lines) < 1 || len(params.Lines) > 3 {
		return nil, fmt.Errorf("请选择1~3条线路")
	}

	seen := make(map[string]bool)
	for _, lp := range params.Lines {
		if seen[lp.LineNo] {
			return nil, fmt.Errorf("不能重复选择线路: %s", lp.LineNo)
		}
		seen[lp.LineNo] = true
	}

	lineResults := make([]models.SimResult, 0, len(params.Lines))
	lineParamsMap := make(map[string]models.SingleLineSimParams)
	date := params.Date
	for _, lp := range params.Lines {
		lineParamsMap[lp.LineNo] = lp
	}

	selectedLineSet := make(map[string]bool)
	for _, lp := range params.Lines {
		selectedLineSet[lp.LineNo] = true
	}

	for _, lp := range params.Lines {
		sp := &models.SimParams{
			LineNo:          lp.LineNo,
			PeakInterval:    lp.PeakInterval,
			OffPeakInterval: lp.OffPeakInterval,
			StationDelta:    lp.StationDelta,
			Date:            date,
		}
		sr, err := s.RunSimulation(sp)
		if err != nil {
			return nil, err
		}
		lineResults = append(lineResults, *sr)
	}

	if len(lineResults) >= 2 {
		s.recalcCrossStationImpacts(lineResults, selectedLineSet, date, lineParamsMap)
	}

	conflicts := s.calcSharedVehicleConflicts(lineResults, selectedLineSet, date)

	mergedAdjImpacts, adjOverloadCount := s.mergeAdjacentImpacts(lineResults, selectedLineSet, date)

	jointOverview := s.buildJointOverview(lineResults, conflicts, mergedAdjImpacts, adjOverloadCount)

	colors := make([]string, len(lineResults))
	for i := range lineResults {
		if i < len(lineColors) {
			colors[i] = lineColors[i]
		} else {
			colors[i] = lineColors[len(lineColors)-1]
		}
	}

	return &models.JointSimResult{
		LineResults:   lineResults,
		JointOverview: jointOverview,
		LineColors:    colors,
	}, nil
}

func (s *SimulationService) buildJointOverview(lineResults []models.SimResult, conflicts []models.SharedVehicleConflict, mergedAdjImpacts []models.AdjLineMergedImpact, adjOverloadCount int) models.JointOverview {
	totalOrigTrips := 0
	totalNewTrips := 0
	totalOrigLoad := 0.0
	totalNewLoad := 0.0

	for _, lr := range lineResults {
		totalOrigTrips += lr.OrigTotalTrips
		totalNewTrips += lr.NewTotalTrips
		totalOrigLoad += lr.OrigKPI.PeakLoadFactor
		totalNewLoad += lr.NewKPI.PeakLoadFactor
	}

	n := float64(len(lineResults))
	avgOrig := 0.0
	avgNew := 0.0
	if n > 0 {
		avgOrig = math.Round(totalOrigLoad/n*10000) / 10000
		avgNew = math.Round(totalNewLoad/n*10000) / 10000
	}

	pct := 0.0
	if totalOrigTrips > 0 {
		pct = math.Round(float64(totalNewTrips-totalOrigTrips)/float64(totalOrigTrips)*10000) / 100
	}

	return models.JointOverview{
		TotalOrigTrips:         totalOrigTrips,
		TotalNewTrips:          totalNewTrips,
		TotalTripsDelta:        totalNewTrips - totalOrigTrips,
		TotalTripsChangePct:    pct,
		AvgOrigLoadFactor:      avgOrig,
		AvgNewLoadFactor:       avgNew,
		AvgLoadFactorDelta:     math.Round((avgNew-avgOrig)*10000) / 10000,
		SharedVehicleConflicts: conflicts,
		HasVehicleConflict:     len(conflicts) > 0,
		MergedAdjacentImpacts:  mergedAdjImpacts,
		AdjacentOverloadCount:  adjOverloadCount,
	}
}

func (s *SimulationService) getVehiclesForLines(lineNos []string, date string) map[string][]string {
	result := make(map[string][]string)
	if len(lineNos) == 0 {
		return result
	}

	placeholders := make([]string, len(lineNos))
	args := make([]interface{}, len(lineNos)+1)
	for i, ln := range lineNos {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = ln
	}
	args[len(lineNos)] = date

	query := fmt.Sprintf(`
		SELECT DISTINCT line_no, vehicle_no
		FROM trips
		WHERE line_no IN (%s) AND trip_date = $%d::date
		ORDER BY line_no, vehicle_no
	`, strings.Join(placeholders, ","), len(lineNos)+1)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var ln, vn string
		if err := rows.Scan(&ln, &vn); err == nil {
			result[ln] = append(result[ln], vn)
		}
	}
	return result
}

func (s *SimulationService) calcSharedVehicleConflicts(lineResults []models.SimResult, selectedLineSet map[string]bool, date string) []models.SharedVehicleConflict {
	lineNos := make([]string, 0, len(lineResults))
	for _, lr := range lineResults {
		lineNos = append(lineNos, lr.LineNo)
	}

	vehicleToLines := make(map[string]map[string]bool)
	vehiclesByLine := s.getVehiclesForLines(lineNos, date)
	for ln, vs := range vehiclesByLine {
		for _, v := range vs {
			if vehicleToLines[v] == nil {
				vehicleToLines[v] = make(map[string]bool)
			}
			vehicleToLines[v][ln] = true
		}
	}

	peakTripsByLine := make(map[string]int)
	for _, lr := range lineResults {
		peakTripsByLine[lr.LineNo] = lr.NewPeakTrips
	}

	vehicleDailyCapacity := s.estimateVehicleDailyPeakCapacity(date)

	conflicts := []models.SharedVehicleConflict{}
	for vn, lineSet := range vehicleToLines {
		if len(lineSet) < 2 {
			continue
		}
		totalPeak := 0
		involvedLines := []string{}
		for ln := range lineSet {
			totalPeak += peakTripsByLine[ln]
			involvedLines = append(involvedLines, ln)
		}
		sort.Strings(involvedLines)
		capLimit := vehicleDailyCapacity
		if totalPeak > capLimit {
			conflicts = append(conflicts, models.SharedVehicleConflict{
				VehicleNo:      vn,
				InvolvedLines:  involvedLines,
				TotalPeakTrips: totalPeak,
				CapacityLimit:  capLimit,
			})
		}
	}

	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].VehicleNo < conflicts[j].VehicleNo
	})
	return conflicts
}

func (s *SimulationService) estimateVehicleDailyPeakCapacity(date string) int {
	peakHours := s.getPeakHours()
	minPeakInterval := MinInterval
	if minPeakInterval <= 0 {
		minPeakInterval = 3
	}
	if peakHours <= 0 {
		return 40
	}
	cap := int(math.Ceil(peakHours * 60 / float64(minPeakInterval)))
	if cap < 20 {
		cap = 20
	}
	return cap
}

type lineFlowContext struct {
	flows          []stationFlow
	removeIndexes  []int
	removedStations map[string]int
	sharedStations map[string]bool
	stationDelta   int
}

func (s *SimulationService) recalcCrossStationImpacts(lineResults []models.SimResult, selectedSet map[string]bool, date string, lineParamsMap map[string]models.SingleLineSimParams) {
	lineCtxMap := make(map[string]*lineFlowContext)

	for i := range lineResults {
		lr := &lineResults[i]
		lp := lineParamsMap[lr.LineNo]
		flows := s.getAggregatedStationFlows(lr.LineNo, date)
		shared := s.getSharedStationNames(lr.LineNo)
		removeCount := 0
		if lp.StationDelta < 0 {
			removeCount = -lp.StationDelta
		}
		indexes := s.getRemoveIndexes(flows, removeCount, shared)

		removed := make(map[string]int)
		for _, idx := range indexes {
			if idx >= 0 && idx < len(flows) {
				removed[flows[idx].StationName] = flows[idx].BoardCount
			}
		}

		lineCtxMap[lr.LineNo] = &lineFlowContext{
			flows:           flows,
			removeIndexes:   indexes,
			removedStations: removed,
			sharedStations:  shared,
			stationDelta:    lp.StationDelta,
		}
	}

	for i := range lineResults {
		lr := &lineResults[i]
		ctx := lineCtxMap[lr.LineNo]
		if ctx.stationDelta >= 0 {
			continue
		}
		if len(ctx.removedStations) == 0 {
			continue
		}

		newAdjImpacts := s.calcAdjacentImpactsWithCross(lr.LineNo, ctx, lineCtxMap, selectedSet, date)
		lr.AdjacentImpacts = newAdjImpacts
	}
}

func (s *SimulationService) calcAdjacentImpactsWithCross(targetLineNo string, ctx *lineFlowContext, allCtx map[string]*lineFlowContext, selectedSet map[string]bool, date string) []models.AdjLineImpact {
	result := []models.AdjLineImpact{}

	if len(ctx.removedStations) == 0 {
		return result
	}

	allLines := s.getAllLines()
	for _, line := range allLines {
		if line.LineNo == targetLineNo {
			continue
		}

		adjCtx := allCtx[line.LineNo]
		adjIsSelected := selectedSet[line.LineNo]

		sharedWithAdj := s.getSharedStations(line.LineNo, ctx.removedStations)
		if len(sharedWithAdj) == 0 {
			continue
		}

		var origPeakLoad float64
		origPeakLoad, _ = s.metricsSvc.calcDailyLoadFactor(line.LineNo, date)

		adjAlsoRemovedSet := make(map[string]bool)
		if adjIsSelected && adjCtx != nil {
			for st := range adjCtx.removedStations {
				adjAlsoRemovedSet[st] = true
			}
		}

		transferPax := 0
		for _, st := range sharedWithAdj {
			boardCnt, ok := ctx.removedStations[st]
			if !ok {
				continue
			}
			baseRatio := 0.3
			if adjIsSelected {
				if adjAlsoRemovedSet[st] {
					baseRatio = 0.0
				} else {
					baseRatio = 0.5
				}
			}
			transferPax += int(float64(boardCnt) * baseRatio)
		}

		adjTrips := s.getTripCount(line.LineNo, date)
		additionalLoad := 0.0
		if adjTrips > 0 {
			additionalLoad = float64(transferPax) / float64(adjTrips) / RatedCapacity
		}

		newPeakLoad := origPeakLoad + additionalLoad
		if newPeakLoad > 1.0 {
			newPeakLoad = 1.0
		}
		overloadRisk := newPeakLoad > 0.9

		result = append(result, models.AdjLineImpact{
			LineNo:         line.LineNo,
			LineName:       line.LineName,
			OrigPeakLoad:   origPeakLoad,
			NewPeakLoad:    math.Round(newPeakLoad*10000) / 10000,
			LoadIncrement:  math.Round(additionalLoad*10000) / 10000,
			OverloadRisk:   overloadRisk,
			SharedStations: sharedWithAdj,
		})
	}

	return result
}

func (s *SimulationService) mergeAdjacentImpacts(lineResults []models.SimResult, selectedSet map[string]bool, date string) ([]models.AdjLineMergedImpact, int) {
	type accumulator struct {
		lineNo         string
		lineName       string
		origPeakLoad   float64
		totalIncrement float64
		sharedSet      map[string]bool
		affectedBySet  map[string]bool
	}

	accMap := make(map[string]*accumulator)

	for _, lr := range lineResults {
		for _, adj := range lr.AdjacentImpacts {
			a, ok := accMap[adj.LineNo]
			if !ok {
				a = &accumulator{
					lineNo:        adj.LineNo,
					lineName:      adj.LineName,
					origPeakLoad:  adj.OrigPeakLoad,
					sharedSet:     make(map[string]bool),
					affectedBySet: make(map[string]bool),
				}
				accMap[adj.LineNo] = a
			}
			a.totalIncrement += adj.LoadIncrement
			for _, st := range adj.SharedStations {
				a.sharedSet[st] = true
			}
			a.affectedBySet[lr.LineNo] = true
		}
	}

	merged := []models.AdjLineMergedImpact{}
	overloadCount := 0
	for _, a := range accMap {
		sharedList := []string{}
		for st := range a.sharedSet {
			sharedList = append(sharedList, st)
		}
		sort.Strings(sharedList)

		affectedList := []string{}
		for ln := range a.affectedBySet {
			affectedList = append(affectedList, ln)
		}
		sort.Strings(affectedList)

		newLoad := a.origPeakLoad + a.totalIncrement
		if newLoad > 1.0 {
			newLoad = 1.0
		}
		overloadRisk := newLoad > 0.9
		if overloadRisk {
			overloadCount++
		}

		merged = append(merged, models.AdjLineMergedImpact{
			LineNo:         a.lineNo,
			LineName:       a.lineName,
			OrigPeakLoad:   math.Round(a.origPeakLoad*10000) / 10000,
			NewPeakLoad:    math.Round(newLoad*10000) / 10000,
			LoadIncrement:  math.Round(a.totalIncrement*10000) / 10000,
			OverloadRisk:   overloadRisk,
			SharedStations: sharedList,
			AffectedBy:     affectedList,
		})
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].LoadIncrement > merged[j].LoadIncrement
	})

	return merged, overloadCount
}
