package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"howtv-server/models"
)

// テスト用のデータベースとルーターをセットアップ
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	// テストモードに設定
	gin.SetMode(gin.TestMode)

	// インメモリSQLiteデータベースを使用
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("テスト用データベースの接続に失敗しました: " + err.Error())
	}

	// テーブルの自動マイグレーション
	db.AutoMigrate(&models.Company{}, &models.JobPosting{}, &models.Position{})

	// コントローラーでDBを使用できるようにする
	DB = db

	// テスト用ルーターの作成
	r := gin.Default()
	return r, db
}

// テスト用のJobPostingとPositionを作成
func createTestData(db *gorm.DB) (models.JobPosting, error) {
	// ポジション作成
	frontendPos := models.Position{Name: "フロントエンドエンジニア"}
	if err := db.FirstOrCreate(&frontendPos, models.Position{Name: "フロントエンドエンジニア"}).Error; err != nil {
		return models.JobPosting{}, err
	}

	// JobPosting作成
	job := models.JobPosting{
		UUID:           uuid.New(),
		Title:          "テスト求人",
		Description:    "テスト説明",
		Requirements:   "テスト要件",
		SalaryRange:    "600万〜800万円",
		Location:       "東京都渋谷区",
		EmploymentType: "正社員",
		Status:         "公開中",
	}

	if err := db.Create(&job).Error; err != nil {
		return models.JobPosting{}, err
	}

	// ポジションを関連付け
	if err := db.Model(&job).Association("Positions").Append(&frontendPos); err != nil {
		return models.JobPosting{}, err
	}

	return job, nil
}

// TestGetJobPostings はGetJobPostingsハンドラーをテストします
func TestGetJobPostings(t *testing.T) {
	r, db := setupTestRouter()

	// テスト用のデータを作成
	job, err := createTestData(db)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// ルーターにハンドラーを登録
	r.GET("/api/v1/jobs", GetJobPostings)

	// リクエストの実行
	req, _ := http.NewRequest("GET", "/api/v1/jobs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// レスポンスのステータスコードを確認
	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスのボディを確認
	var response []models.JobPosting
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("レスポンスのパースに失敗しました: %v", err)
	}

	// 少なくとも1件のジョブポスティングが返されていることを確認
	assert.True(t, len(response) > 0)

	// 作成したジョブポスティングが含まれていることを確認
	found := false
	for _, j := range response {
		if j.UUID == job.UUID {
			found = true
			assert.Equal(t, "テスト求人", j.Title)
			break
		}
	}
	assert.True(t, found, "作成したジョブポスティングが応答に含まれていません")
}

// TestGetJobPosting はGetJobPostingハンドラーをテストします
func TestGetJobPosting(t *testing.T) {
	r, db := setupTestRouter()

	// テスト用のデータを作成
	job, err := createTestData(db)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// ルーターにハンドラーを登録
	r.GET("/api/v1/jobs/:uuid", GetJobPosting)

	// リクエストの実行
	req, _ := http.NewRequest("GET", "/api/v1/jobs/"+job.UUID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// レスポンスのステータスコードを確認
	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスのボディを確認
	var response models.JobPosting
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("レスポンスのパースに失敗しました: %v", err)
	}

	// 取得したジョブポスティングが正しいことを確認
	assert.Equal(t, job.UUID, response.UUID)
	assert.Equal(t, "テスト求人", response.Title)
	assert.Equal(t, "東京都渋谷区", response.Location)
}

// TestCreateJobPosting はCreateJobPostingハンドラーをテストします
func TestCreateJobPosting(t *testing.T) {
	r, db := setupTestRouter()

	// ポジションを作成
	frontendPos := models.Position{Name: "フロントエンドエンジニア"}
	if err := db.FirstOrCreate(&frontendPos, models.Position{Name: "フロントエンドエンジニア"}).Error; err != nil {
		t.Fatalf("Positionの作成に失敗しました: %v", err)
	}

	// リクエストボディの作成
	reqBody := struct {
		Title          string    `json:"title"`
		Description    string    `json:"description"`
		Requirements   string    `json:"requirements"`
		SalaryRange    string    `json:"salary_range"`
		Location       string    `json:"location"`
		EmploymentType string    `json:"employment_type"`
		PostingDate    time.Time `json:"posting_date"`
		ClosingDate    time.Time `json:"closing_date"`
		Status         string    `json:"status"`
		PositionIDs    []uint    `json:"position_ids"`
	}{
		Title:          "新規テスト求人",
		Description:    "新規テスト説明",
		Requirements:   "新規テスト要件",
		SalaryRange:    "700万〜900万円",
		Location:       "フルリモート",
		EmploymentType: "正社員",
		PostingDate:    time.Now(),
		ClosingDate:    time.Now().AddDate(0, 1, 0),
		Status:         "公開中",
		PositionIDs:    []uint{frontendPos.ID},
	}

	body, _ := json.Marshal(reqBody)

	// ルーターにハンドラーを登録
	r.POST("/api/v1/jobs", CreateJobPosting)

	// リクエストの実行
	req, _ := http.NewRequest("POST", "/api/v1/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// レスポンスのステータスコードを確認
	assert.Equal(t, http.StatusCreated, w.Code)

	// レスポンスのボディを確認
	var response models.JobPosting
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("レスポンスのパースに失敗しました: %v", err)
	}

	// 作成されたジョブポスティングが正しいことを確認
	assert.Equal(t, "新規テスト求人", response.Title)
	assert.Equal(t, "フルリモート", response.Location)
	assert.NotEqual(t, uuid.Nil, response.UUID)

	// ポジションが関連付けられていることを確認
	assert.Equal(t, 1, len(response.Positions))
	assert.Equal(t, "フロントエンドエンジニア", response.Positions[0].Name)
}
