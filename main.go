package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	// CORS設定
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

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

		// デバッグエンドポイント（開発環境のみ）
		if gin.Mode() != gin.ReleaseMode {
			debug := v1.Group("/debug")
			{
				debug.GET("/stats", controllers.DataStats)
				debug.POST("/fix-positions", controllers.FixMissingPositions)
				debug.POST("/reseed", controllers.ResetAndReseed)
			}
		}
	}

	return r
}

func initDatabase() {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "test.db"
	}

	log.Printf("データベースファイル: %s", dbPath)

	// データベース接続
	controllers.DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}

	// スキーママイグレーション
	log.Println("データベースマイグレーションを開始します...")
	err = controllers.DB.AutoMigrate(&models.Company{}, &models.JobPosting{}, &models.Position{})
	if err != nil {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}
	log.Println("マイグレーション完了")

	// データベースの内容を確認（再シードするかどうか判断するため）
	var positionCount, jobCount int64
	controllers.DB.Model(&models.Position{}).Count(&positionCount)
	controllers.DB.Model(&models.JobPosting{}).Count(&jobCount)

	log.Printf("データベース状態: ポジション %d件, 求人 %d件", positionCount, jobCount)

	// モックデータファイルのパスを確認
	mockDataPath := "mockdata.txt"
	if _, err := os.Stat(mockDataPath); os.IsNotExist(err) {
		// カレントディレクトリの絶対パスを取得
		cwd, _ := os.Getwd()
		// 代替パスを試す
		alternativePaths := []string{
			"mockdata.txt",
			filepath.Join(cwd, "mockdata.txt"),
			filepath.Join(cwd, "..", "mockdata.txt"),
		}

		for _, path := range alternativePaths {
			if _, err := os.Stat(path); err == nil {
				mockDataPath = path
				log.Printf("モックデータファイルを見つけました: %s", mockDataPath)
				break
			}
		}
	}

	// データがない、または明示的に再シードが要求された場合（環境変数RESEED=true）
	needsSeeding := positionCount == 0 || jobCount == 0 || os.Getenv("RESEED") == "true"

	if needsSeeding {
		log.Println("データベースをシードします...")

		// 既存のポジション関連をリセット（重複防止）
		if jobCount > 0 {
			log.Println("既存のポジション関連をリセットします...")
			if err := scripts.ResetPositionAssociations(); err != nil {
				log.Printf("警告: ポジション関連のリセットに失敗しました: %v", err)
			}
		}

		// データベースのシード
		if err := scripts.SeedDatabase(mockDataPath); err != nil {
			log.Printf("警告: データベースのシードに失敗しました: %v", err)
		} else {
			log.Println("データベースのシードに成功しました")
		}

		// シード後のデータ確認
		controllers.DB.Model(&models.Position{}).Count(&positionCount)
		controllers.DB.Model(&models.JobPosting{}).Count(&jobCount)
		log.Printf("シード後のデータ状態: ポジション %d件, 求人 %d件", positionCount, jobCount)

		// 求人とポジションの関連を確認
		var associationCount int64
		controllers.DB.Table("job_positions").Count(&associationCount)
		log.Printf("求人とポジションの関連: %d件", associationCount)
	} else {
		log.Println("データベースは既に初期化されています。シードはスキップします。")
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
