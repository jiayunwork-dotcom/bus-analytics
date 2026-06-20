package services

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"bus-analytics/internal/db"
	"bus-analytics/internal/models"
)

type UploadService struct{}

func NewUploadService() *UploadService {
	return &UploadService{}
}

func (s *UploadService) ParseAndValidate(file *multipart.FileHeader, dataType string) (*models.UploadResponse, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	reader := csv.NewReader(bufio.NewReader(src))
	reader.Comma = ','
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return &models.UploadResponse{
			Success: false,
			Message: "CSV文件至少需要包含表头和一行数据",
		}, nil
	}

	headers := records[0]
	for i, h := range headers {
		headers[i] = strings.TrimSpace(strings.ToLower(h))
	}

	var requiredFields []string
	switch dataType {
	case "routes":
		requiredFields = []string{"线路编号", "线路名", "起讫站", "全程公里数", "站点数量", "票价"}
	case "trips":
		requiredFields = []string{"线路编号", "日期", "班次号", "实发时间", "计划时间", "车辆编号", "方向"}
	case "flows":
		requiredFields = []string{"线路编号", "日期", "班次号", "站点序号", "站点名", "上客人数", "下客人数", "刷卡人数"}
	case "mileages":
		requiredFields = []string{"车辆编号", "日期", "总里程", "营运里程"}
	default:
		return &models.UploadResponse{
			Success: false,
			Message: "不支持的数据类型: " + dataType,
		}, nil
	}

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	errors := []models.ValidationError{}
	inserted := 0

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for rowIdx := 1; rowIdx < len(records); rowIdx++ {
		row := records[rowIdx]
		missing := []string{}
		for _, req := range requiredFields {
			idx, ok := headerMap[strings.ToLower(req)]
			if !ok || idx >= len(row) || strings.TrimSpace(row[idx]) == "" {
				missing = append(missing, req)
			}
		}

		if len(missing) > 0 {
			errors = append(errors, models.ValidationError{
				Row:    rowIdx + 1,
				Fields: missing,
			})
			continue
		}

		switch dataType {
		case "routes":
			if err := s.insertRoute(tx, row, headerMap); err != nil {
				errors = append(errors, models.ValidationError{Row: rowIdx + 1, Fields: []string{err.Error()}})
				continue
			}
		case "trips":
			if err := s.insertTrip(tx, row, headerMap); err != nil {
				errors = append(errors, models.ValidationError{Row: rowIdx + 1, Fields: []string{err.Error()}})
				continue
			}
		case "flows":
			if err := s.insertFlow(tx, row, headerMap); err != nil {
				errors = append(errors, models.ValidationError{Row: rowIdx + 1, Fields: []string{err.Error()}})
				continue
			}
		case "mileages":
			if err := s.insertMileage(tx, row, headerMap); err != nil {
				errors = append(errors, models.ValidationError{Row: rowIdx + 1, Fields: []string{err.Error()}})
				continue
			}
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &models.UploadResponse{
		Success:       true,
		Message:       fmt.Sprintf("数据导入完成，成功%d条，失败%d条", inserted, len(errors)),
		DataType:      dataType,
		InsertedCount: inserted,
		Errors:        errors,
	}, nil
}

func (s *UploadService) insertRoute(tx interface{ Exec(string, ...interface{}) (interface{}, error) }, row []string, hm map[string]int) error {
	lineNo := strings.TrimSpace(row[hm["线路编号"]])
	lineName := strings.TrimSpace(row[hm["线路名"]])
	stations := strings.Split(strings.TrimSpace(row[hm["起讫站"]]), "-")
	if len(stations) < 2 {
		stations = strings.Split(strings.TrimSpace(row[hm["起讫站"]]), "—")
	}
	if len(stations) < 2 {
		return fmt.Errorf("起讫站格式错误")
	}
	totalKm, _ := strconv.ParseFloat(strings.TrimSpace(row[hm["全程公里数"]]), 64)
	stationCount, _ := strconv.Atoi(strings.TrimSpace(row[hm["站点数量"]]))
	fare, _ := strconv.ParseFloat(strings.TrimSpace(row[hm["票价"]]), 64)

	var straightDist *float64
	if idx, ok := hm["直线距离"]; ok && idx < len(row) && strings.TrimSpace(row[idx]) != "" {
		d, _ := strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
		straightDist = &d
	}

	_, err := db.DB.Exec(`INSERT INTO routes (line_no, line_name, start_station, end_station, total_km, station_count, fare, straight_line_distance_km)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (line_no) DO UPDATE SET
			line_name = EXCLUDED.line_name,
			start_station = EXCLUDED.start_station,
			end_station = EXCLUDED.end_station,
			total_km = EXCLUDED.total_km,
			station_count = EXCLUDED.station_count,
			fare = EXCLUDED.fare,
			straight_line_distance_km = COALESCE(EXCLUDED.straight_line_distance_km, routes.straight_line_distance_km)`,
		lineNo, lineName, stations[0], stations[len(stations)-1], totalKm, stationCount, fare, straightDist)
	return err
}

func (s *UploadService) insertTrip(tx interface{}, row []string, hm map[string]int) error {
	lineNo := strings.TrimSpace(row[hm["线路编号"]])
	dateStr := strings.TrimSpace(row[hm["日期"]])
	tripNo := strings.TrimSpace(row[hm["班次号"]])
	actualTime := strings.TrimSpace(row[hm["实发时间"]])
	plannedTime := strings.TrimSpace(row[hm["计划时间"]])
	vehicleNo := strings.TrimSpace(row[hm["车辆编号"]])

	var driverName string
	if idx, ok := hm["驾驶员"]; ok && idx < len(row) {
		driverName = strings.TrimSpace(row[idx])
	}

	direction := 1
	if idx, ok := hm["方向"]; ok && idx < len(row) {
		d := strings.TrimSpace(row[idx])
		if d == "下行" || d == "2" || d == "down" {
			direction = 2
		}
	}

	_, err := db.DB.Exec(`INSERT INTO trips (line_no, trip_date, trip_no, actual_departure_time, planned_departure_time, vehicle_no, driver_name, direction)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING`,
		lineNo, dateStr, tripNo, actualTime, plannedTime, vehicleNo, driverName, direction)
	return err
}

func (s *UploadService) insertFlow(tx interface{}, row []string, hm map[string]int) error {
	lineNo := strings.TrimSpace(row[hm["线路编号"]])
	dateStr := strings.TrimSpace(row[hm["日期"]])
	tripNo := strings.TrimSpace(row[hm["班次号"]])
	stationSeq, _ := strconv.Atoi(strings.TrimSpace(row[hm["站点序号"]]))
	stationName := strings.TrimSpace(row[hm["站点名"]])
	board, _ := strconv.Atoi(strings.TrimSpace(row[hm["上客人数"]]))
	alight, _ := strconv.Atoi(strings.TrimSpace(row[hm["下客人数"]]))
	card, _ := strconv.Atoi(strings.TrimSpace(row[hm["刷卡人数"]]))

	var cardID string
	if idx, ok := hm["刷卡人id"]; ok && idx < len(row) {
		cardID = strings.TrimSpace(row[idx])
	}

	_, err := db.DB.Exec(`INSERT INTO station_flows (line_no, flow_date, trip_no, station_seq, station_name, board_count, alight_count, card_count, card_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		lineNo, dateStr, tripNo, stationSeq, stationName, board, alight, card, cardID)
	return err
}

func (s *UploadService) insertMileage(tx interface{}, row []string, hm map[string]int) error {
	vehicleNo := strings.TrimSpace(row[hm["车辆编号"]])
	dateStr := strings.TrimSpace(row[hm["日期"]])
	totalKm, _ := strconv.ParseFloat(strings.TrimSpace(row[hm["总里程"]]), 64)
	operatingKm, _ := strconv.ParseFloat(strings.TrimSpace(row[hm["营运里程"]]), 64)

	_, err := db.DB.Exec(`INSERT INTO vehicle_mileages (vehicle_no, mileage_date, total_km, operating_km)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`,
		vehicleNo, dateStr, totalKm, operatingKm)
	return err
}

func (s *UploadService) GetDataSummary() (*models.DataSummary, error) {
	summary := &models.DataSummary{}

	var minDate, maxDate *string
	db.DB.QueryRow(`SELECT MIN(trip_date)::text, MAX(trip_date)::text FROM trips`).Scan(&minDate, &maxDate)
	if minDate != nil && maxDate != nil {
		summary.DateRange = []string{*minDate, *maxDate}
	}

	db.DB.QueryRow(`SELECT COUNT(DISTINCT line_no) FROM routes`).Scan(&summary.LineCount)
	db.DB.QueryRow(`SELECT COUNT(*) FROM trips`).Scan(&summary.TripCount)
	db.DB.QueryRow(`SELECT COUNT(DISTINCT vehicle_no) FROM trips`).Scan(&summary.VehicleCount)
	db.DB.QueryRow(`SELECT COALESCE(SUM(board_count), 0) FROM station_flows`).Scan(&summary.TotalPassengers)

	return summary, nil
}

func ParseTime(t string) (time.Time, error) {
	formats := []string{"15:04", "15:04:05", "03:04 PM", "2006-01-02 15:04:05"}
	for _, f := range formats {
		if parsed, err := time.Parse(f, t); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", t)
}
