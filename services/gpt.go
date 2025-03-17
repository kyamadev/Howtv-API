package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"howtv-server/models"

	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client *openai.Client
}

type RoadmapResponse struct {
	Roadmap  string `json:"roadmap"`
	Skills   string `json:"skills"`
	Timeline string `json:"timeline"`
}

func NewOpenAIService(apiKey string) *OpenAIService {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	client := openai.NewClient(apiKey)
	return &OpenAIService{
		client: client,
	}
}

func (s *OpenAIService) GenerateCareerRoadmap(job *models.JobPosting, questionType string) (*RoadmapResponse, error) {
	var positionNames []string
	for _, pos := range job.Positions {
		positionNames = append(positionNames, pos.Name)
	}

	// 質問タイプに応じたプロンプトを生成
	prompt := generatePromptByQuestionType(job, positionNames, questionType)

	// ChatGPT APIにリクエスト
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたはキャリア開発とスキルアップに関する専門家です。特定の求人に応募するために必要なスキルセットとロードマップを提供してください。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return nil, fmt.Errorf("ロードマップの生成中にエラーが発生しました: %w", err)
	}

	// レスポンスを処理
	content := resp.Choices[0].Message.Content
	roadmap := parseRoadmapResponse(content)

	return roadmap, nil
}

func generatePromptByQuestionType(job *models.JobPosting, positionNames []string, questionType string) string {
	jobInfo := fmt.Sprintf(`
職種: %s
役職: %s
説明: %s
必要条件: %s
勤務地: %s
雇用形態: %s
`,
		job.Title,
		strings.Join(positionNames, ", "),
		job.Description,
		job.Requirements,
		job.Location,
		job.EmploymentType,
	)

	var specificPrompt string
	switch questionType {
	case "skills":
		specificPrompt = fmt.Sprintf(`
以下の求人情報に基づいて、この職種に必要なスキルと知識について詳細に解説してください:
%s

この回答には以下の要素を含めてください:
1. 必須の技術スキル（プログラミング言語、フレームワーク、ツールなど）の詳細説明
2. あると有利な周辺技術・知識
3. 技術以外の重要なスキル（コミュニケーション能力、問題解決能力など）
4. 各スキルの重要度とスキルレベルの目安

具体的で実践的な説明をお願いします。回答は箇条書きを適宜使用し、見やすく構造化してください。
`, jobInfo)

	case "learning":
		specificPrompt = fmt.Sprintf(`
以下の求人情報に基づいて、必要なスキルと知識を効率的に習得するための具体的な学習方法を提案してください:
%s

この回答には以下の要素を含めてください:
1. 各技術スキルの効果的な学習リソース（オンラインコース、書籍、チュートリアルなど）
2. 初心者から応募レベルまでの具体的な学習ステップ
3. 推奨の学習プロジェクト
4. 学習の進捗を測定する方法
5. 学習にかかる推定時間

具体的かつ実践的な学習プランを提案してください。回答は箇条書きを適宜使用し、見やすく構造化してください。
`, jobInfo)

	case "career":
		specificPrompt = fmt.Sprintf(`
以下の求人情報に基づいて、この職種のキャリアパスと将来の発展可能性について解説してください:
%s

この回答には以下の要素を含めてください:
1. この職種に就いた後のキャリアステップと成長の道筋
2. 次のキャリアレベルに進むために必要なスキルと経験
3. 5年後、10年後のキャリア展望
4. この分野での専門性を深めるための方向性とオプション
5. 関連する他の職種への転向可能性

中長期的なキャリア展望について具体的に解説してください。回答は箇条書きを適宜使用し、見やすく構造化してください。
`, jobInfo)

	default: // general
		specificPrompt = fmt.Sprintf(`
以下の求人情報に基づいて、この職種に必要なスキルと知識を習得するためのロードマップを作成してください:
%s

このロードマップには以下の要素を含めてください:
1. 必要なプログラミング言語と技術スキルの習得方法
2. おすすめの学習リソース（オンラインコース、書籍、チュートリアルなど）
3. スキルレベル別の学習タイムライン（初心者、中級者、応募準備完了）
4. 必要に応じて英語力向上のための勉強方法

回答は日本語で、具体的かつ実践的なアドバイスをお願いします。
`, jobInfo)
	}

	return specificPrompt
}

func parseRoadmapResponse(content string) *RoadmapResponse {
	// 単純化のため、全体のコンテンツをロードマップとして返す
	// 実際のアプリケーションでは、より構造化されたパースが必要かも
	return &RoadmapResponse{
		Roadmap:  content,
		Skills:   "", // APIレスポンスから適切にパースする場合は、この部分を実装
		Timeline: "", // APIレスポンスから適切にパースする場合は、この部分を実装
	}
}
