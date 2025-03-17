package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"howtv-server/config"
	"howtv-server/controllers"
	"howtv-server/models"
	"howtv-server/scripts"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // 開発環境
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(config))

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Job Postings
		v1.GET("/jobs", controllers.GetJobPostings)
		v1.GET("/jobs/:uuid", controllers.GetJobPosting)
		v1.POST("/jobs", controllers.CreateJobPosting)
		v1.PUT("/jobs/:uuid", controllers.UpdateJobPosting)
		v1.DELETE("/jobs/:uuid", controllers.DeleteJobPosting)

		// Positions
		v1.GET("/positions", controllers.GetPositions)
		v1.POST("/positions", controllers.CreatePosition)
		v1.POST("/jobs/:uuid/positions", controllers.AssignPositionsToJob)

		// Roadmap Generation
		v1.GET("/jobs/:uuid/roadmap", controllers.GenerateRoadmap)
	}

	return r
}

func initDatabase() {
	var err error
	controllers.DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	controllers.DB.AutoMigrate(&models.Company{}, &models.JobPosting{}, &models.Position{})

	// Seed the database with mock data
	if err := scripts.SeedDatabase("mockdata.txt"); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}
}

func main() {
	// 環境変数から設定を読み込む
	config.LoadConfig()

	// Initialize database
	initDatabase()

	// Setup router
	r := setupRouter()

	// PORT環境変数を確認
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // デフォルトのポート
	}

	// Listen and serve
	log.Printf("サーバーを起動しました: 0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
