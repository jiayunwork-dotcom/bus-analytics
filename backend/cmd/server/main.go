package main

import (
	"bus-analytics/internal/config"
	"bus-analytics/internal/db"
	"bus-analytics/internal/handlers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	h := handlers.NewHandler()

	api := r.Group("/api")
	{
		api.POST("/upload", h.UploadData)
		api.GET("/summary", h.GetDataSummary)
		api.GET("/lines", h.GetAllLines)
		api.GET("/vehicles", h.GetAllVehicles)
		api.GET("/date-range", h.GetDateRange)

		api.GET("/metrics/lines", h.GetLineEfficiencies)
		api.GET("/metrics/lines/:lineNo/trend", h.GetLineDailyTrend)

		api.GET("/flow/:lineNo/section", h.GetSectionFlow)
		api.GET("/flow/:lineNo/hourly", h.GetHourlyDistribution)
		api.GET("/flow/:lineNo/tidal", h.CheckTidalPattern)
		api.GET("/flow/:lineNo/od", h.InferOD)

		api.GET("/schedule/:lineNo/optimize", h.OptimizeSchedule)

		api.GET("/vehicles/utilization", h.GetVehicleUtilizations)
		api.GET("/vehicles/:vehicleNo/gantt", h.GetVehicleGantt)

		api.GET("/network/metrics", h.GetNetworkMetrics)

		api.POST("/compare/lines", h.CompareLines)
		api.GET("/compare/lines/:lineNo/historical", h.GetLineHistoricalTrend)
		api.GET("/compare/health-scores", h.GetLineHealthScores)

		api.POST("/report/export", h.ExportReport)

		api.POST("/simulation/line", h.RunLineSimulation)
		api.POST("/simulation/joint", h.RunJointSimulation)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
}
