package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"howtv-server/config"
	"howtv-server/controllers"
	"howtv-server/models"
)

var testRouter *gin.Engine
var testDB *gorm.DB

// TestMain はテスト全体のセットアップとクリーンアップを行います
func TestMain(m *testing.M) {
	// テスト用の環境変数を設定
	os.Setenv("OPENAI_API_KEY", "test-api-key-for-testing")
	os.Setenv("PORT", "8081")

	// テストの前の準備
	setup()

	// テストの実行
	code := m.Run()

	// テスト後のクリーンアップ
	teardown()

	// 終了コードを返す
	os.Exit(code)
}

// setup はテスト環境をセットアップします
// setup はテスト環境をセットアップします
func setup() {
	// 現在のディレクトリを表示（デバッグ用）
	cwd, _ := os.Getwd()
	fmt.Printf("テスト実行時のカレントディレクトリ: %s\n", cwd)

	// .envファイルの存在確認
	if _, err := os.Stat(".env"); err == nil {
		fmt.Println("カレントディレクトリに.envファイルがあります")
	} else {
		fmt.Printf("カレントディレクトリに.envファイルがありません: %v\n", err)

		// 上位ディレクトリも確認
		if _, err := os.Stat("../.env"); err == nil {
			fmt.Println("親ディレクトリに.envファイルがあります")
		}
	}

	// GinをTestモードに設定
	gin.SetMode(gin.TestMode)

	// テスト用のDBを作成
	var err error
	testDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("テスト用データベースの接続に失敗しました: " + err.Error())
	}

	// マイグレーション
	testDB.AutoMigrate(&models.Company{}, &models.JobPosting{}, &models.Position{})

	// コントローラーにDBをセット
	controllers.DB = testDB

	// 設定を読み込む
	config.LoadConfig()

	// テスト用のデータを作成
	seedTestData()

	// ルーターのセットアップ
	testRouter = setupTestRouter()
}

// teardown はテスト後のクリーンアップを行います
func teardown() {
	// 必要に応じてリソースを解放
}

// setupTestRouter はテスト用のルーターをセットアップします
func setupTestRouter() *gin.Engine {
	r := gin.Default()

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

		// Roadmap Generation (モックモードでのみテスト)
		// v1.GET("/jobs/:uuid/roadmap", controllers.GenerateRoadmap)
	}

	return r
}

// seedTestData はテスト用のデータを作成します
func seedTestData() {
	// ポジションを作成
	positions := []models.Position{
		{Name: "フロントエンドエンジニア"},
		{Name: "バックエンドエンジニア"},
		{Name: "フルスタックエンジニア"},
	}

	for i := range positions {
		testDB.FirstOrCreate(&positions[i], models.Position{Name: positions[i].Name})
	}

	// 会社を作成
	company := models.Company{
		Name:    "テスト株式会社",
		Address: "東京都渋谷区",
		Website: "https://test.co.jp",
	}
	testDB.FirstOrCreate(&company, models.Company{Name: "テスト株式会社"})

	// 求人を作成
	job := models.JobPosting{
		UUID:           uuid.New(),
		CompanyID:      company.UUID,
		Title:          "テストエンジニア",
		Description:    "テスト説明文",
		Requirements:   "テスト要件",
		SalaryRange:    "600万〜800万円",
		Location:       "東京都渋谷区",
		EmploymentType: "正社員",
		Status:         "公開中",
	}

	// 求人を保存
	result := testDB.Create(&job)
	if result.Error != nil {
		fmt.Printf("テスト求人の作成に失敗しました: %v\n", result.Error)
		return
	}

	// ポジションを関連付け
	testDB.Model(&job).Association("Positions").Append(&positions[0], &positions[1])
}

// TestGetAllJobs は全ての求人取得APIをテストします
func TestGetAllJobs(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/jobs", nil)
	testRouter.ServeHTTP(w, req)

	// ステータスコードを確認
	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスをパース
	var response []models.JobPosting
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 少なくとも1件の求人が返されていることを確認
	assert.True(t, len(response) > 0)
	assert.Equal(t, "テストエンジニア", response[0].Title)
}

// TestGetSingleJob は特定の求人取得APIをテストします
func TestGetSingleJob(t *testing.T) {
	// 最初の求人を取得
	var job models.JobPosting
	testDB.First(&job)

	// APIリクエスト
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/jobs/"+job.UUID.String(), nil)
	testRouter.ServeHTTP(w, req)

	// ステータスコードを確認
	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスをパース
	var response models.JobPosting
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 取得した求人が正しいことを確認
	assert.Equal(t, job.UUID, response.UUID)
	assert.Equal(t, "テストエンジニア", response.Title)

	// ポジションが含まれていることを確認
	assert.Equal(t, 2, len(response.Positions))
}

// TestGetPositions はポジション一覧取得APIをテストします
func TestGetPositions(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/positions", nil)
	testRouter.ServeHTTP(w, req)

	// ステータスコードを確認
	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスをパース
	var response []models.Position
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 少なくとも3つのポジションが返されていることを確認
	assert.True(t, len(response) >= 3)

	// 期待されるポジションが含まれていることを確認
	positionNames := make(map[string]bool)
	for _, pos := range response {
		positionNames[pos.Name] = true
	}

	assert.True(t, positionNames["フロントエンドエンジニア"])
	assert.True(t, positionNames["バックエンドエンジニア"])
	assert.True(t, positionNames["フルスタックエンジニア"])
}

// TestInvalidJobUUID は不正なUUIDでの求人取得APIをテストします
func TestInvalidJobUUID(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/jobs/invalid-uuid", nil)
	testRouter.ServeHTTP(w, req)

	// ステータスコードを確認 (400 Bad Request)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// レスポンスにエラーメッセージが含まれていることを確認
	// コントローラーから返される実際のメッセージと一致させる
	assert.Contains(t, w.Body.String(), "Invalid UUID format")
}

// TestNonExistentJob は存在しない求人取得APIをテストします
func TestNonExistentJob(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/jobs/"+uuid.New().String(), nil)
	testRouter.ServeHTTP(w, req)

	// ステータスコードを確認 (404 Not Found)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// レスポンスにエラーメッセージが含まれていることを確認
	// コントローラーから返される実際のメッセージと一致させる
	assert.Contains(t, w.Body.String(), "Job posting not found")
}
