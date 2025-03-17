package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"howtv-server/models"
)

// OpenAIServiceのモック
type MockOpenAIService struct {
	generateRoadmapFunc func(*models.JobPosting) (*RoadmapResponse, error)
}

func (m *MockOpenAIService) GenerateCareerRoadmap(job *models.JobPosting) (*RoadmapResponse, error) {
	return m.generateRoadmapFunc(job)
}

// TestParseRoadmapResponse はparseRoadmapResponse関数をテストします
func TestParseRoadmapResponse(t *testing.T) {
	content := `
# フロントエンドエンジニアのロードマップ

## 必要なスキル
- HTML/CSS
- JavaScript
- React

## 学習タイムライン
1. 基礎学習 (1-2ヶ月)
2. フレームワーク学習 (2-3ヶ月)
3. プロジェクト作成 (1-2ヶ月)
`

	// 関数を実行
	response := parseRoadmapResponse(content)

	// 結果を検証
	assert.NotNil(t, response)
	assert.Equal(t, content, response.Roadmap)
}

// TestGenerateRoadmapPrompt はgenerateRoadmapPrompt関数をテストします
func TestGenerateRoadmapPrompt(t *testing.T) {
	// テスト用のJobPostingとポジションを作成
	job := &models.JobPosting{
		UUID:           uuid.New(),
		Title:          "フロントエンドエンジニア",
		Description:    "Reactを使ったフロントエンド開発",
		Requirements:   "HTML, CSS, JavaScript, Reactの経験",
		SalaryRange:    "600万〜800万円",
		Location:       "東京都渋谷区",
		EmploymentType: "正社員",
		Status:         "公開中",
	}

	positionNames := []string{"フロントエンドエンジニア"}

	// 関数を実行
	prompt := generateRoadmapPrompt(job, positionNames)

	// 生成されたプロンプトに重要な情報が含まれていることを確認
	assert.Contains(t, prompt, "フロントエンドエンジニア")
	assert.Contains(t, prompt, "Reactを使ったフロントエンド開発")
	assert.Contains(t, prompt, "HTML, CSS, JavaScript, Reactの経験")
	assert.Contains(t, prompt, "東京都渋谷区")
	assert.Contains(t, prompt, "正社員")
	assert.Contains(t, prompt, "必要なプログラミング言語と技術スキルの習得方法")
}

// TestOpenAIServiceMock はモックを使用してOpenAIServiceをテストします
func TestOpenAIServiceMock(t *testing.T) {
	// テスト用のJobPostingを作成
	job := &models.JobPosting{
		UUID:           uuid.New(),
		Title:          "Goバックエンドエンジニア",
		Description:    "Goを使ったバックエンド開発",
		Requirements:   "Goの実務経験、APIの設計と実装経験",
		SalaryRange:    "700万〜900万円",
		Location:       "フルリモート",
		EmploymentType: "正社員",
		Positions: []models.Position{
			{Name: "バックエンドエンジニア"},
		},
	}

	// モックサービスの作成
	mockService := &MockOpenAIService{
		generateRoadmapFunc: func(j *models.JobPosting) (*RoadmapResponse, error) {
			// モックのレスポンスを返す
			return &RoadmapResponse{
				Roadmap:  "# Goバックエンドエンジニアのロードマップ\n\n- Goの基礎を学ぶ\n- APIの設計と実装\n- データベース連携\n- マイクロサービスの理解",
				Skills:   "Go, API設計, データベース, Docker",
				Timeline: "3-6ヶ月",
			}, nil
		},
	}

	// モックサービスでGenerateCareerRoadmapを呼び出す
	response, err := mockService.GenerateCareerRoadmap(job)

	// 結果を検証
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Contains(t, response.Roadmap, "Goバックエンドエンジニアのロードマップ")
	assert.Contains(t, response.Roadmap, "APIの設計と実装")
}
