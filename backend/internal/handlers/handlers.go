package handlers

import (
	"bus-analytics/internal/db"
	"bus-analytics/internal/services"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func dbQuery(q string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(q, args...)
}

type Handler struct {
	uploadSvc     *services.UploadService
	metricsSvc    *services.MetricsService
	flowSvc       *services.FlowAnalysisService
	scheduleSvc   *services.ScheduleOptimizationService
	vehicleSvc    *services.VehicleService
	networkSvc    *services.NetworkService
	comparisonSvc *services.ComparisonService
	reportSvc     *services.ReportService
}

func NewHandler() *Handler {
	return &Handler{
		uploadSvc:     services.NewUploadService(),
		metricsSvc:    services.NewMetricsService(),
		flowSvc:       services.NewFlowAnalysisService(),
		scheduleSvc:   services.NewScheduleOptimizationService(),
		vehicleSvc:    services.NewVehicleService(),
		networkSvc:    services.NewNetworkService(),
		comparisonSvc: services.NewComparisonService(),
		reportSvc:     services.NewReportService(),
	}
}

func (h *Handler) UploadData(c *gin.Context) {
	dataType := c.PostForm("data_type")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请上传文件"})
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请上传CSV格式文件"})
		return
	}

	result, err := h.uploadSvc.ParseAndValidate(file, dataType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetDataSummary(c *gin.Context) {
	summary, err := h.uploadSvc.GetDataSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) GetLineEfficiencies(c *gin.Context) {
	effs, err := h.metricsSvc.GetAllLineEfficiencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, effs)
}

func (h *Handler) GetLineDailyTrend(c *gin.Context) {
	lineNo := c.Param("lineNo")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	trend, err := h.metricsSvc.GetLineDailyTrend(lineNo, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

func (h *Handler) GetSectionFlow(c *gin.Context) {
	lineNo := c.Param("lineNo")
	date := c.Query("date")
	flow, maxPax, maxDir, err := h.flowSvc.GetSectionFlow(lineNo, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"sections":            flow,
		"max_passengers":      maxPax,
		"max_direction":       maxDir,
	})
}

func (h *Handler) GetHourlyDistribution(c *gin.Context) {
	lineNo := c.Param("lineNo")
	date := c.Query("date")
	dist, err := h.flowSvc.GetHourlyDistribution(lineNo, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dist)
}

func (h *Handler) CheckTidalPattern(c *gin.Context) {
	lineNo := c.Param("lineNo")
	date := c.Query("date")
	isTidal, emptyRate, direction, err := h.flowSvc.CheckTidalPattern(lineNo, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"is_tidal":          isTidal,
		"empty_rate":        emptyRate,
		"tidal_direction":   direction,
	})
}

func (h *Handler) InferOD(c *gin.Context) {
	lineNo := c.Param("lineNo")
	date := c.Query("date")
	successRate, ods, err := h.flowSvc.InferOD(lineNo, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success_rate": successRate,
		"od_pairs":     ods,
	})
}

func (h *Handler) OptimizeSchedule(c *gin.Context) {
	lineNo := c.Param("lineNo")
	date := c.Query("date")
	totalVehicles, _ := strconv.Atoi(c.DefaultQuery("total_vehicles", "20"))
	result, err := h.scheduleSvc.OptimizeSchedule(lineNo, totalVehicles, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetVehicleUtilizations(c *gin.Context) {
	date := c.Query("date")
	result, lowCount, highCount, avgTrips, err := h.vehicleSvc.GetVehicleUtilizations(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"vehicles":   result,
		"low_count":  lowCount,
		"high_count": highCount,
		"avg_trips":  avgTrips,
	})
}

func (h *Handler) GetVehicleGantt(c *gin.Context) {
	vehicleNo := c.Param("vehicleNo")
	date := c.Query("date")
	gantt, err := h.vehicleSvc.GetVehicleGantt(vehicleNo, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gantt)
}

func (h *Handler) GetNetworkMetrics(c *gin.Context) {
	cityArea, _ := strconv.ParseFloat(c.DefaultQuery("city_area", "1000"), 64)
	uniqueLineKm, _ := strconv.ParseFloat(c.DefaultQuery("unique_line_km", "500"), 64)
	metrics, err := h.networkSvc.GetNetworkMetrics(cityArea, uniqueLineKm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) CompareLines(c *gin.Context) {
	var req struct {
		LineNos []string `json:"line_nos" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择2-5条线路"})
		return
	}
	if len(req.LineNos) < 2 || len(req.LineNos) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择2-5条线路"})
		return
	}
	result, err := h.comparisonSvc.CompareLines(req.LineNos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetLineHistoricalTrend(c *gin.Context) {
	lineNo := c.Param("lineNo")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	granularity := c.DefaultQuery("granularity", "week")
	trend, err := h.comparisonSvc.GetLineHistoricalTrend(lineNo, startDate, endDate, granularity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

func (h *Handler) ExportReport(c *gin.Context) {
	var req struct {
		LineNos []string `json:"line_nos" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择线路"})
		return
	}
	metrics, err := h.comparisonSvc.CompareLines(req.LineNos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	pdfBytes, err := h.reportSvc.GenerateMonthlyReport(req.LineNos, metrics, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=monthly_report.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) GetAllLines(c *gin.Context) {
	type lineInfo struct {
		LineNo   string `json:"line_no"`
		LineName string `json:"line_name"`
	}
	var lines []lineInfo
	rows, err := dbQuery("SELECT line_no, line_name FROM routes ORDER BY line_no")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var l lineInfo
		if err := rows.Scan(&l.LineNo, &l.LineName); err == nil {
			lines = append(lines, l)
		}
	}
	c.JSON(http.StatusOK, lines)
}

func (h *Handler) GetAllVehicles(c *gin.Context) {
	var vehicles []string
	rows, err := dbQuery("SELECT DISTINCT vehicle_no FROM trips ORDER BY vehicle_no")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err == nil {
			vehicles = append(vehicles, v)
		}
	}
	c.JSON(http.StatusOK, vehicles)
}

func (h *Handler) GetDateRange(c *gin.Context) {
	var minDate, maxDate *string
	db.DB.QueryRow("SELECT MIN(trip_date)::text, MAX(trip_date)::text FROM trips").Scan(&minDate, &maxDate)
	result := make(map[string]interface{})
	if minDate != nil {
		result["min_date"] = *minDate
	}
	if maxDate != nil {
		result["max_date"] = *maxDate
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetLineHealthScores(c *gin.Context) {
	scores, err := h.comparisonSvc.GetLineHealthScores()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scores)
}
