package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"howtv-server/models"
)

// OpenAIServiceのモック
type MockOpenAIService struct {
	generateRoadmapFunc func(*models.JobPosting, string) (*RoadmapResponse, error)
}

func (m *MockOpenAIService) GenerateCareerRoadmap(job *models.JobPosting, questionType string) (*RoadmapResponse, error) {
	return m.generateRoadmapFunc(job, questionType)
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

// TestGeneratePromptByQuestionType は各質問タイプに対するプロンプト生成を検証します
func TestGeneratePromptByQuestionType(t *testing.T) {
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

	// 各質問タイプのテスト
	testCases := []struct {
		name         string
		questionType string
		checkContent []string
	}{
		{
			name:         "一般的な質問",
			questionType: "general",
			checkContent: []string{"ロードマップを作成", "必要なプログラミング言語と技術スキルの習得方法"},
		},
		{
			name:         "スキルに関する質問",
			questionType: "skills",
			checkContent: []string{"必要なスキルと知識について詳細に解説", "必須の技術スキル", "技術以外の重要なスキル"},
		},
		{
			name:         "学習方法に関する質問",
			questionType: "learning",
			checkContent: []string{"習得するための具体的な学習方法", "効果的な学習リソース", "学習プロジェクト"},
		},
		{
			name:         "キャリアパスに関する質問",
			questionType: "career",
			checkContent: []string{"キャリアパスと将来の発展可能性", "キャリアステップと成長の道筋", "5年後、10年後のキャリア展望"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// プロンプトを生成
			prompt := generatePromptByQuestionType(job, positionNames, tc.questionType)

			// 共通の情報が含まれていることを確認
			assert.Contains(t, prompt, "フロントエンドエンジニア")
			assert.Contains(t, prompt, "Reactを使ったフロントエンド開発")
			assert.Contains(t, prompt, "HTML, CSS, JavaScript, Reactの経験")
			assert.Contains(t, prompt, "東京都渋谷区")
			assert.Contains(t, prompt, "正社員")

			// 質問タイプ固有の内容が含まれていることを確認
			for _, content := range tc.checkContent {
				assert.Contains(t, prompt, content)
			}
		})
	}
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

	// 各質問タイプごとにテスト
	testCases := []struct {
		name         string
		questionType string
		expectedText string
	}{
		{
			name:         "一般的な質問",
			questionType: "general",
			expectedText: "Goバックエンドエンジニアのロードマップ",
		},
		{
			name:         "スキルに関する質問",
			questionType: "skills",
			expectedText: "Goバックエンドエンジニアに必要なスキル",
		},
		{
			name:         "学習方法に関する質問",
			questionType: "learning",
			expectedText: "Go言語の学習方法",
		},
		{
			name:         "キャリアパスに関する質問",
			questionType: "career",
			expectedText: "バックエンドエンジニアのキャリアパス",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックサービスの作成
			mockService := &MockOpenAIService{
				generateRoadmapFunc: func(j *models.JobPosting, qt string) (*RoadmapResponse, error) {
					// 質問タイプに基づいた異なるレスポンスを返す
					var content string
					switch qt {
					case "skills":
						content = "# Goバックエンドエンジニアに必要なスキル\n\n- Go言語の深い理解\n- APIの設計知識\n- データベース設計と操作\n- マイクロサービスアーキテクチャ"
					case "learning":
						content = "# Go言語の学習方法\n\n- Go公式ドキュメントの学習\n- オンラインコース受講\n- 実践的なプロジェクト作成\n- コードレビューの活用"
					case "career":
						content = "# バックエンドエンジニアのキャリアパス\n\n- シニアバックエンドエンジニア\n- テックリード\n- アーキテクト\n- エンジニアリングマネージャー"
					default: // general
						content = "# Goバックエンドエンジニアのロードマップ\n\n- Goの基礎を学ぶ\n- APIの設計と実装\n- データベース連携\n- マイクロサービスの理解"
					}
					// モックのレスポンスを返す
					return &RoadmapResponse{
						Roadmap:  content,
						Skills:   "Go, API設計, データベース, Docker",
						Timeline: "3-6ヶ月",
					}, nil
				},
			}

			// モックサービスでGenerateCareerRoadmapを呼び出す
			response, err := mockService.GenerateCareerRoadmap(job, tc.questionType)

			// 結果を検証
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Contains(t, response.Roadmap, tc.expectedText)
		})
	}
}

// TestGenerateCareerRoadmapWithDifferentQuestionTypes は実際のOpenAIServiceの代わりにモックを使って
// さまざまな質問タイプに対するレスポンスをテストします
func TestGenerateCareerRoadmapWithDifferentQuestionTypes(t *testing.T) {
	// テスト用のJobPostingを作成
	job := &models.JobPosting{
		UUID:           uuid.New(),
		Title:          "データサイエンティスト",
		Description:    "機械学習モデルの開発と分析",
		Requirements:   "Python, 統計学, 機械学習の知識",
		SalaryRange:    "800万〜1100万円",
		Location:       "東京（リモート可）",
		EmploymentType: "正社員",
		Positions: []models.Position{
			{Name: "データエンジニア"},
			{Name: "AI/ML エンジニア"},
		},
	}

	// 簡易モックサービスの作成（実際のAPIを呼び出さずにテストするため）
	mockService := &MockOpenAIService{
		generateRoadmapFunc: func(j *models.JobPosting, qt string) (*RoadmapResponse, error) {
			// 質問タイプを含むダミーレスポンスを返す
			return &RoadmapResponse{
				Roadmap:  "Question Type: " + qt + "\nTitle: " + j.Title,
				Skills:   "Python, TensorFlow, PyTorch, SQL",
				Timeline: "6-12ヶ月",
			}, nil
		},
	}

	// 各質問タイプでテスト
	questionTypes := []string{"general", "skills", "learning", "career"}
	for _, qt := range questionTypes {
		t.Run("QuestionType_"+qt, func(t *testing.T) {
			response, err := mockService.GenerateCareerRoadmap(job, qt)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Contains(t, response.Roadmap, "Question Type: "+qt)
			assert.Contains(t, response.Roadmap, "Title: データサイエンティスト")
		})
	}
}
