package services

import (
	"bus-analytics/internal/models"
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

type ReportService struct{}

func NewReportService() *ReportService {
	return &ReportService{}
}

func (s *ReportService) GenerateMonthlyReport(lineNos []string, metrics []models.LineEfficiency, sectionData string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "B", 16)
	pdf.AddPage()

	pdf.Cell(40, 10, "Bus Operations Monthly Report")
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Line Efficiency Metrics:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	headers := []string{"Line No.", "Name", "Passenger Intensity", "Peak Load", "Off-Peak Load", "Speed(km/h)", "On-time Rate(%)"}
	w := []float64{20, 35, 30, 22, 25, 22, 25}
	for i, h := range headers {
		pdf.CellFormat(w[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	for _, m := range metrics {
		pdf.CellFormat(w[0], 6, m.LineNo, "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[1], 6, truncate(m.LineName, 12), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[2], 6, fmt.Sprintf("%.2f", m.PassengerIntensity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[3], 6, fmt.Sprintf("%.2f", m.PeakLoadFactor), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[4], 6, fmt.Sprintf("%.2f", m.OffPeakLoadFactor), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[5], 6, fmt.Sprintf("%.2f", m.OperatingSpeed), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[6], 6, fmt.Sprintf("%.2f", m.OnTimeRate), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Schedule Optimization Recommendations:")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 6, "Optimize dispatch intervals during peak hours to reduce average passenger waiting time. Adjust based on passenger flow distribution patterns.", "", "L", false)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
