package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"sort"
)

type FlowAnalysisService struct{}

func NewFlowAnalysisService() *FlowAnalysisService {
	return &FlowAnalysisService{}
}

func (s *FlowAnalysisService) GetSectionFlow(lineNo, date string) (map[int][]models.SectionFlow, int, int, error) {
	result := make(map[int][]models.SectionFlow)
	maxFlow := 0
	maxDirection := 1

	for direction := 1; direction <= 2; direction++ {
		flows, dirMax := s.calcDirectionSectionFlow(lineNo, date, direction)
		result[direction] = flows
		if dirMax > maxFlow {
			maxFlow = dirMax
			maxDirection = direction
		}
	}

	return result, maxFlow, maxDirection, nil
}

func (s *FlowAnalysisService) calcDirectionSectionFlow(lineNo, date string, direction int) ([]models.SectionFlow, int) {
	rows, err := db.DB.Query(`
		SELECT trip_no
		FROM trips
		WHERE line_no = $1 AND trip_date = $2 AND direction = $3
	`, lineNo, date, direction)
	if err != nil {
		return []models.SectionFlow{}, 0
	}
	defer rows.Close()

	var tripNos []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tripNos = append(tripNos, t)
		}
	}

	sectionAccum := make(map[int]map[string]interface{})
	stationNames := make(map[int]string)
	maxPassengers := 0

	for _, tripNo := range tripNos {
		stationRows, err := db.DB.Query(`
			SELECT station_seq, station_name, board_count, alight_count
			FROM station_flows
			WHERE line_no = $1 AND trip_no = $2 AND flow_date = $3
			ORDER BY station_seq
		`, lineNo, tripNo, date)
		if err != nil {
			continue
		}

		onBoard := 0
		prevSeq := 0
		prevName := ""
		for stationRows.Next() {
			var seq int
			var name string
			var board, alight int
			if err := stationRows.Scan(&seq, &name, &board, &alight); err != nil {
				continue
			}
			stationNames[seq] = name
			onBoard = onBoard - alight + board

			if prevSeq > 0 {
				key := prevSeq
				if sectionAccum[key] == nil {
					sectionAccum[key] = map[string]interface{}{
						"from": prevName,
						"to":   name,
						"sum":  0,
						"cnt":  0,
					}
				}
				m := sectionAccum[key]
				m["sum"] = m["sum"].(int) + onBoard
				m["cnt"] = m["cnt"].(int) + 1
			}
			prevSeq = seq
			prevName = name
		}
		stationRows.Close()
	}

	var sections []models.SectionFlow
	for seq, m := range sectionAccum {
		sum := m["sum"].(int)
		cnt := m["cnt"].(int)
		avg := 0
		if cnt > 0 {
			avg = sum / cnt
		}
		if avg > maxPassengers {
			maxPassengers = avg
		}
		sections = append(sections, models.SectionFlow{
			StationSeq:  seq,
			FromStation: m["from"].(string),
			ToStation:   m["to"].(string),
			Passengers:  avg,
			Direction:   direction,
		})
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].StationSeq < sections[j].StationSeq
	})

	return sections, maxPassengers
}

func (s *FlowAnalysisService) GetHourlyDistribution(lineNo, date string) ([]models.HourlyDistribution, error) {
	rows, err := db.DB.Query(`
		SELECT EXTRACT(HOUR FROM t.actual_departure_time) as hour,
			SUM(s.board_count)
		FROM station_flows s
		INNER JOIN trips t ON s.line_no = t.line_no AND s.trip_no = t.trip_no AND s.flow_date = t.trip_date
		WHERE s.line_no = $1 AND s.flow_date = $2
		GROUP BY hour
		ORDER BY hour
	`, lineNo, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.HourlyDistribution, 24)
	for i := range result {
		result[i] = models.HourlyDistribution{Hour: i}
	}

	for rows.Next() {
		var hour sql.NullFloat64
		var cnt sql.NullInt64
		if err := rows.Scan(&hour, &cnt); err == nil && hour.Valid {
			idx := int(hour.Float64)
			if idx >= 0 && idx < 24 {
				result[idx].Boarders = int(cnt.Int64)
			}
		}
	}

	return result, nil
}

func (s *FlowAnalysisService) CheckTidalPattern(lineNo, date string) (bool, float64, string, error) {
	morningUp := s.getDirectionHourlyBoard(lineNo, date, 1, 7, 9)
	morningDown := s.getDirectionHourlyBoard(lineNo, date, 2, 7, 9)
	eveningUp := s.getDirectionHourlyBoard(lineNo, date, 1, 17, 19)
	eveningDown := s.getDirectionHourlyBoard(lineNo, date, 2, 17, 19)

	isTidal := false
	direction := ""
	if morningUp > morningDown*2 || eveningDown > eveningUp*2 {
		isTidal = true
		direction = "上行早高峰/下行晚高峰"
	}
	if morningDown > morningUp*2 || eveningUp > eveningDown*2 {
		isTidal = true
		direction = "下行早高峰/上行晚高峰"
	}

	emptyRate := s.calcOppositeEmptyRate(lineNo, date)
	return isTidal, emptyRate, direction, nil
}

func (s *FlowAnalysisService) getDirectionHourlyBoard(lineNo, date string, direction, startHour, endHour int) int {
	var total sql.NullInt64
	db.DB.QueryRow(`
		SELECT SUM(s.board_count)
		FROM station_flows s
		INNER JOIN trips t ON s.line_no = t.line_no AND s.trip_no = t.trip_no AND s.flow_date = t.trip_date
		WHERE s.line_no = $1 AND s.flow_date = $2 AND t.direction = $3
			AND EXTRACT(HOUR FROM t.actual_departure_time) >= $4
			AND EXTRACT(HOUR FROM t.actual_departure_time) < $5
	`, lineNo, date, direction, startHour, endHour).Scan(&total)
	if total.Valid {
		return int(total.Int64)
	}
	return 0
}

func (s *FlowAnalysisService) calcOppositeEmptyRate(lineNo, date string) float64 {
	var totalTrips, lowLoadTrips sql.NullInt64
	db.DB.QueryRow(`
		WITH trip_loads AS (
			SELECT t.trip_no, t.direction,
				(SELECT MAX(running_count) FROM (
					SELECT 
						SUM(board_count - alight_count) OVER (ORDER BY s.station_seq) as running_count
					FROM station_flows s
					WHERE s.line_no = t.line_no AND s.trip_no = t.trip_no AND s.flow_date = t.trip_date
				) sub) as max_load
			FROM trips t
			WHERE t.line_no = $1 AND t.trip_date = $2
		)
		SELECT COUNT(*), COUNT(*) FILTER (WHERE max_load < 16)
		FROM trip_loads
	`, lineNo, date).Scan(&totalTrips, &lowLoadTrips)

	if !totalTrips.Valid || totalTrips.Int64 == 0 {
		return 0
	}
	return float64(lowLoadTrips.Int64) / float64(totalTrips.Int64) * 100
}

func (s *FlowAnalysisService) InferOD(lineNo, date string) (float64, []models.ODPair, error) {
	rows, err := db.DB.Query(`
		SELECT card_id, station_seq, station_name, flow_date::text
		FROM station_flows
		WHERE line_no = $1 AND flow_date = $2 AND card_id IS NOT NULL AND card_id != ''
		ORDER BY card_id, station_seq
	`, lineNo, date)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	type cardTrip struct {
		cardID      string
		stationSeq  int
		stationName string
		date        string
	}

	tripsByCard := make(map[string][]cardTrip)
	totalCards := 0
	for rows.Next() {
		var ct cardTrip
		var cardID sql.NullString
		if err := rows.Scan(&cardID, &ct.stationSeq, &ct.stationName, &ct.date); err == nil && cardID.Valid {
			ct.cardID = cardID.String
			if _, exists := tripsByCard[ct.cardID]; !exists {
				totalCards++
			}
			tripsByCard[ct.cardID] = append(tripsByCard[ct.cardID], ct)
		}
	}

	odPairs := make(map[string]int)
	successCount := 0

	for _, trips := range tripsByCard {
		if len(trips) >= 2 {
			origin := trips[0].stationName
			dest := trips[1].stationName
			key := origin + "→" + dest
			odPairs[key]++
			successCount++
		}
	}

	successRate := 0.0
	if totalCards > 0 {
		successRate = float64(successCount) / float64(totalCards) * 100
	}

	var ods []models.ODPair
	for k, v := range odPairs {
		parts := splitODKey(k)
		ods = append(ods, models.ODPair{
			Origin:      parts[0],
			Destination: parts[1],
			Count:       v,
		})
	}

	sort.Slice(ods, func(i, j int) bool {
		return ods[i].Count > ods[j].Count
	})
	if len(ods) > 20 {
		ods = ods[:20]
	}

	return successRate, ods, nil
}

func splitODKey(key string) []string {
	for i := 0; i < len(key); i++ {
		if i+3 <= len(key) && key[i:i+3] == "→" {
			return []string{key[:i], key[i+3:]}
		}
	}
	return []string{key, ""}
}
