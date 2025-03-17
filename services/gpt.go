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

func (s *OpenAIService) GenerateCareerRoadmap(job *models.JobPosting) (*RoadmapResponse, error) {
	var positionNames []string
	for _, pos := range job.Positions {
		positionNames = append(positionNames, pos.Name)
	}

	// プロンプトを作成
	prompt := generateRoadmapPrompt(job, positionNames)

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

func generateRoadmapPrompt(job *models.JobPosting, positionNames []string) string {
	prompt := fmt.Sprintf(`
以下の求人情報に基づいて、この職種に必要なスキルと知識を習得するためのロードマップを作成してください:

職種: %s
役職: %s
説明: %s
必要条件: %s
勤務地: %s
雇用形態: %s

このロードマップには以下の要素を含めてください:
1. 必要なプログラミング言語と技術スキルの習得方法
2. おすすめの学習リソース（オンラインコース、書籍、チュートリアルなど）
3. スキルレベル別の学習タイムライン（初心者、中級者、応募準備完了）
4. 必要に応じて英語力向上のための勉強方法

回答は日本語で、具体的かつ実践的なアドバイスをお願いします。
`,
		job.Title,
		strings.Join(positionNames, ", "),
		job.Description,
		job.Requirements,
		job.Location,
		job.EmploymentType,
	)

	return prompt
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
