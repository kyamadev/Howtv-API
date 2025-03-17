package models

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// テスト用のデータベースをセットアップ
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("テスト用データベースの接続に失敗しました: " + err.Error())
	}

	// テーブルの自動マイグレーション
	db.AutoMigrate(&JobPosting{}, &Position{}, &Company{})

	return db
}

// TestJobPostingCreate はJobPostingの作成をテストします
func TestJobPostingCreate(t *testing.T) {
	db := setupTestDB()

	// テスト用のデータを作成
	job := JobPosting{
		Title:          "テストタイトル",
		Description:    "テスト説明",
		Requirements:   "テスト要件",
		SalaryRange:    "600万〜800万円",
		Location:       "東京都渋谷区",
		EmploymentType: "正社員",
		Status:         "公開中",
	}

	// データベースに保存
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("JobPostingの作成に失敗しました: %v", err)
	}

	// UUIDが自動生成されているか確認
	if job.UUID == uuid.Nil {
		t.Error("UUIDが自動生成されていません")
	}

	// 保存されたデータを取得して検証
	var savedJob JobPosting
	if err := db.First(&savedJob, "uuid = ?", job.UUID).Error; err != nil {
		t.Fatalf("JobPostingの取得に失敗しました: %v", err)
	}

	// データが正しく保存されているか確認
	if savedJob.Title != "テストタイトル" {
		t.Errorf("期待されるタイトル「テストタイトル」に対して「%s」が保存されています", savedJob.Title)
	}
}

// TestJobPostingWithPositions はJobPostingとPositionの関連をテストします
func TestJobPostingWithPositions(t *testing.T) {
	db := setupTestDB()

	// ポジションを作成
	frontendPos := Position{Name: "フロントエンドエンジニア"}
	backendPos := Position{Name: "バックエンドエンジニア"}

	if err := db.Create(&frontendPos).Error; err != nil {
		t.Fatalf("Positionの作成に失敗しました: %v", err)
	}
	if err := db.Create(&backendPos).Error; err != nil {
		t.Fatalf("Positionの作成に失敗しました: %v", err)
	}

	// JobPostingを作成
	job := JobPosting{
		Title:          "フルスタックエンジニア",
		Description:    "フロントエンドとバックエンドの両方を担当",
		Requirements:   "React, Goの経験",
		SalaryRange:    "700万〜900万円",
		Location:       "フルリモート",
		EmploymentType: "正社員",
		Status:         "公開中",
		Positions:      []Position{frontendPos, backendPos},
	}

	// データベースに保存
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("JobPostingとPositionsの関連付けに失敗しました: %v", err)
	}

	// 関連するポジションを取得して検証
	var savedJob JobPosting
	if err := db.Preload("Positions").First(&savedJob, "uuid = ?", job.UUID).Error; err != nil {
		t.Fatalf("関連するPositionsの取得に失敗しました: %v", err)
	}

	// 関連するポジションの数を確認
	if len(savedJob.Positions) != 2 {
		t.Errorf("期待されるPosition数「2」に対して「%d」が関連付けられています", len(savedJob.Positions))
	}

	// 関連するポジションの名前を確認
	positionNames := make(map[string]bool)
	for _, pos := range savedJob.Positions {
		positionNames[pos.Name] = true
	}

	if !positionNames["フロントエンドエンジニア"] {
		t.Error("「フロントエンドエンジニア」ポジションが関連付けられていません")
	}
	if !positionNames["バックエンドエンジニア"] {
		t.Error("「バックエンドエンジニア」ポジションが関連付けられていません")
	}
}
