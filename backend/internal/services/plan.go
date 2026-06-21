package services

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type PlanService struct{}

func NewPlanService() *PlanService {
	return &PlanService{}
}

func (s *PlanService) CreatePlan(plan *models.SimPlan) (*models.SimPlan, error) {
	paramsJSON, err := json.Marshal(plan.Params)
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %w", err)
	}
	resultJSON, err := json.Marshal(plan.Result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	var id int
	var createdAt time.Time
	err = db.DB.QueryRow(`
		INSERT INTO sim_plans (name, remark, sim_type, lines, params, result)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, plan.Name, plan.Remark, plan.SimType, plan.Lines, paramsJSON, resultJSON).Scan(&id, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("保存方案失败: %w", err)
	}

	plan.ID = id
	plan.CreatedAt = createdAt
	return plan, nil
}

func (s *PlanService) ListPlans(sortOrder string, lineFilter string) ([]models.SimPlanListItem, error) {
	orderBy := "created_at DESC"
	if sortOrder == "asc" {
		orderBy = "created_at ASC"
	}

	query := fmt.Sprintf(`
		SELECT id, name, remark, sim_type, lines, created_at, params, result
		FROM sim_plans
		ORDER BY %s
	`, orderBy)

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询方案列表失败: %w", err)
	}
	defer rows.Close()

	var result []models.SimPlanListItem
	for rows.Next() {
		var item models.SimPlanListItem
		var paramsJSON, resultJSON []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Remark, &item.SimType, &item.Lines, &item.CreatedAt, &paramsJSON, &resultJSON); err != nil {
			continue
		}

		if lineFilter != "" && !strings.Contains(item.Lines, lineFilter) {
			continue
		}

		s.extractKPIFromResult(resultJSON, &item)
		result = append(result, item)
	}
	return result, nil
}

func (s *PlanService) extractKPIFromResult(resultJSON []byte, item *models.SimPlanListItem) {
	if item.SimType == "single" {
		var sr models.SimResult
		if err := json.Unmarshal(resultJSON, &sr); err == nil {
			item.DailyTrips = sr.NewKPI.DailyTrips
			item.PeakLoadFactor = sr.NewKPI.PeakLoadFactor
			item.OperatingSpeed = sr.NewKPI.OperatingSpeed
			item.PassengerIntensity = sr.NewKPI.PassengerIntensity
		}
	} else {
		var jr models.JointSimResult
		if err := json.Unmarshal(resultJSON, &jr); err == nil {
			if len(jr.LineResults) > 0 {
				totalTrips := 0
				totalLoad := 0.0
				totalSpeed := 0.0
				totalIntensity := 0.0
				n := len(jr.LineResults)
				for _, lr := range jr.LineResults {
					totalTrips += lr.NewKPI.DailyTrips
					totalLoad += lr.NewKPI.PeakLoadFactor
					totalSpeed += lr.NewKPI.OperatingSpeed
					totalIntensity += lr.NewKPI.PassengerIntensity
				}
				item.DailyTrips = totalTrips
				item.PeakLoadFactor = totalLoad / float64(n)
				item.OperatingSpeed = totalSpeed / float64(n)
				item.PassengerIntensity = totalIntensity / float64(n)
			}
		}
	}
}

func (s *PlanService) GetPlan(id int) (*models.SimPlan, error) {
	var plan models.SimPlan
	var paramsJSON, resultJSON []byte
	err := db.DB.QueryRow(`
		SELECT id, name, remark, sim_type, lines, params, result, created_at
		FROM sim_plans WHERE id = $1
	`, id).Scan(&plan.ID, &plan.Name, &plan.Remark, &plan.SimType, &plan.Lines, &paramsJSON, &resultJSON, &plan.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("方案不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询方案失败: %w", err)
	}

	var paramsAny, resultAny any
	json.Unmarshal(paramsJSON, &paramsAny)
	json.Unmarshal(resultJSON, &resultAny)
	plan.Params = paramsAny
	plan.Result = resultAny
	return &plan, nil
}

func (s *PlanService) DeletePlan(id int) error {
	res, err := db.DB.Exec(`DELETE FROM sim_plans WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("删除方案失败: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("方案不存在")
	}
	return nil
}

func (s *PlanService) RenamePlan(id int, name string) error {
	res, err := db.DB.Exec(`UPDATE sim_plans SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return fmt.Errorf("重命名方案失败: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("方案不存在")
	}
	return nil
}

func (s *PlanService) ComparePlans(planIDs []int) (*models.SimPlanCompareResult, error) {
	if len(planIDs) < 2 || len(planIDs) > 4 {
		return nil, fmt.Errorf("请选择2到4份方案进行对比")
	}

	plans := make([]models.SimPlan, 0, len(planIDs))
	for _, pid := range planIDs {
		p, err := s.GetPlan(pid)
		if err != nil {
			return nil, fmt.Errorf("方案%d: %w", pid, err)
		}
		plans = append(plans, *p)
	}

	compareItems := make([]models.SimPlanCompareItem, 0, len(plans))
	jointOverviews := make([]models.JointOverviewItem, 0)

	for _, p := range plans {
		item := models.SimPlanCompareItem{
			ID:        p.ID,
			Name:      p.Name,
			SimType:   p.SimType,
			Lines:     p.Lines,
			CreatedAt: p.CreatedAt,
		}

		resultJSON, _ := json.Marshal(p.Result)

		if p.SimType == "single" {
			var sr models.SimResult
			if err := json.Unmarshal(resultJSON, &sr); err == nil {
				item.KPI = models.PlanKPI{
					DailyTrips:         sr.NewKPI.DailyTrips,
					PeakLoadFactor:     sr.NewKPI.PeakLoadFactor,
					OperatingSpeed:     sr.NewKPI.OperatingSpeed,
					PassengerIntensity: sr.NewKPI.PassengerIntensity,
				}
			}
		} else {
			var jr models.JointSimResult
			if err := json.Unmarshal(resultJSON, &jr); err == nil {
				if len(jr.LineResults) > 0 {
					totalTrips := 0
					totalLoad := 0.0
					totalSpeed := 0.0
					totalIntensity := 0.0
					n := len(jr.LineResults)
					for _, lr := range jr.LineResults {
						totalTrips += lr.NewKPI.DailyTrips
						totalLoad += lr.NewKPI.PeakLoadFactor
						totalSpeed += lr.NewKPI.OperatingSpeed
						totalIntensity += lr.NewKPI.PassengerIntensity
					}
					item.KPI = models.PlanKPI{
						DailyTrips:         totalTrips,
						PeakLoadFactor:     totalLoad / float64(n),
						OperatingSpeed:     totalSpeed / float64(n),
						PassengerIntensity: totalIntensity / float64(n),
					}
				}

				jo := jr.JointOverview
				jointOverviews = append(jointOverviews, models.JointOverviewItem{
					PlanID:              p.ID,
					PlanName:            p.Name,
					TotalOrigTrips:      jo.TotalOrigTrips,
					TotalNewTrips:       jo.TotalNewTrips,
					TotalTripsDelta:     jo.TotalTripsDelta,
					TotalTripsChangePct: jo.TotalTripsChangePct,
					AvgOrigLoadFactor:   jo.AvgOrigLoadFactor,
					AvgNewLoadFactor:    jo.AvgNewLoadFactor,
					AvgLoadFactorDelta:  jo.AvgLoadFactorDelta,
				})
			}
		}

		compareItems = append(compareItems, item)
	}

	paramDiffs := s.buildParamDiffs(plans)
	recommendations := s.buildRecommendations(compareItems)

	return &models.SimPlanCompareResult{
		Plans:          compareItems,
		JointOverviews: jointOverviews,
		ParamDiffs:     paramDiffs,
		Recommendations: recommendations,
	}, nil
}

type planParamEntry struct {
	LineNo          string
	PeakInterval    int
	OffPeakInterval int
	StationDelta    int
}

func (s *PlanService) extractPlanParams(plan models.SimPlan) []planParamEntry {
	paramsJSON, _ := json.Marshal(plan.Params)
	var entries []planParamEntry

	if plan.SimType == "single" {
		var sp models.SimParams
		if err := json.Unmarshal(paramsJSON, &sp); err == nil {
			entries = append(entries, planParamEntry{
				LineNo:          sp.LineNo,
				PeakInterval:    sp.PeakInterval,
				OffPeakInterval: sp.OffPeakInterval,
				StationDelta:    sp.StationDelta,
			})
		}
	} else {
		var jp models.JointSimParams
		if err := json.Unmarshal(paramsJSON, &jp); err == nil {
			for _, lp := range jp.Lines {
				entries = append(entries, planParamEntry{
					LineNo:          lp.LineNo,
					PeakInterval:    lp.PeakInterval,
					OffPeakInterval: lp.OffPeakInterval,
					StationDelta:    lp.StationDelta,
				})
			}
		}
	}
	return entries
}

func (s *PlanService) buildParamDiffs(plans []models.SimPlan) []models.ParamDiffRow {
	allEntries := make([][]planParamEntry, len(plans))
	for i, p := range plans {
		allEntries[i] = s.extractPlanParams(p)
	}

	maxLines := 0
	for _, entries := range allEntries {
		if len(entries) > maxLines {
			maxLines = len(entries)
		}
	}

	var rows []models.ParamDiffRow

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		lineNos := make([]string, len(plans))
		for i, entries := range allEntries {
			if lineIdx < len(entries) {
				lineNos[i] = entries[lineIdx].LineNo
			}
		}

		if lineIdx == 0 || hasVariation(lineNos) {
			label := "线路编号"
			if maxLines > 1 {
				label = fmt.Sprintf("线路%d 编号", lineIdx+1)
			}
			rows = append(rows, buildDiffRow(label, lineNos, plans))
		}

		peakVals := make([]any, len(plans))
		offPeakVals := make([]any, len(plans))
		stationVals := make([]any, len(plans))
		for i, entries := range allEntries {
			if lineIdx < len(entries) {
				peakVals[i] = entries[lineIdx].PeakInterval
				offPeakVals[i] = entries[lineIdx].OffPeakInterval
				stationVals[i] = entries[lineIdx].StationDelta
			}
		}

		prefix := ""
		if maxLines > 1 {
			prefix = fmt.Sprintf("线路%d ", lineIdx+1)
		}
		rows = append(rows, buildDiffRowAny(prefix+"高峰间隔(分)", peakVals, plans))
		rows = append(rows, buildDiffRowAny(prefix+"平峰间隔(分)", offPeakVals, plans))
		rows = append(rows, buildDiffRowAny(prefix+"站点增减数", stationVals, plans))
	}

	return rows
}

func hasVariation(vals []string) bool {
	if len(vals) < 2 {
		return false
	}
	first := vals[0]
	for _, v := range vals[1:] {
		if v != first {
			return true
		}
	}
	return false
}

func buildDiffRow(paramName string, vals []string, plans []models.SimPlan) models.ParamDiffRow {
	same := !hasVariation(vals)
	values := make([]models.ParamDiffValue, len(vals))
	for i, v := range vals {
		values[i] = models.ParamDiffValue{
			PlanID: plans[i].ID,
			Value:  v,
			Same:   same,
		}
	}
	return models.ParamDiffRow{ParamName: paramName, Values: values, Same: same}
}

func buildDiffRowAny(paramName string, vals []any, plans []models.SimPlan) models.ParamDiffRow {
	if len(vals) < 2 {
		return models.ParamDiffRow{ParamName: paramName, Values: nil, Same: true}
	}

	first := fmt.Sprintf("%v", vals[0])
	same := true
	for _, v := range vals[1:] {
		if fmt.Sprintf("%v", v) != first {
			same = false
			break
		}
	}

	values := make([]models.ParamDiffValue, len(vals))
	for i, v := range vals {
		values[i] = models.ParamDiffValue{
			PlanID: plans[i].ID,
			Value:  v,
			Same:   same,
		}
	}
	return models.ParamDiffRow{ParamName: paramName, Values: values, Same: same}
}

func (s *PlanService) buildRecommendations(items []models.SimPlanCompareItem) []models.PlanRecommendation {
	if len(items) < 2 {
		return nil
	}

	type candidate struct {
		idx            int
		peakLoadFactor float64
		operatingSpeed float64
		planID         int
		planName       string
	}

	candidates := make([]candidate, len(items))
	for i, item := range items {
		candidates[i] = candidate{
			idx:            i,
			peakLoadFactor: item.KPI.PeakLoadFactor,
			operatingSpeed: item.KPI.OperatingSpeed,
			planID:         item.ID,
			planName:       item.Name,
		}
	}

	var underLimit []candidate
	for _, c := range candidates {
		if c.peakLoadFactor <= 0.85 {
			underLimit = append(underLimit, c)
		}
	}

	var best *candidate
	var reason string
	if len(underLimit) > 0 {
		sort.Slice(underLimit, func(i, j int) bool {
			return underLimit[i].operatingSpeed > underLimit[j].operatingSpeed
		})
		best = &underLimit[0]
		reason = fmt.Sprintf("满载率%.1f%%≤85%%且营运速度最高(%.2f km/h)", best.peakLoadFactor*100, best.operatingSpeed)
	} else {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].peakLoadFactor < candidates[j].peakLoadFactor
		})
		best = &candidates[0]
		reason = fmt.Sprintf("所有方案满载率均超85%%，取满载率最低(%.1f%%)", best.peakLoadFactor*100)
	}

	return []models.PlanRecommendation{
		{
			PlanID:   best.planID,
			PlanName: best.planName,
			Reason:   reason,
		},
	}
}
