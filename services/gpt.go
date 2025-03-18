package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

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
	// 初期化
	rand.Seed(time.Now().UnixNano())

	var positionNames []string
	for _, pos := range job.Positions {
		positionNames = append(positionNames, pos.Name)
	}

	// 質問タイプに応じたプロンプトを生成
	prompt := generatePromptByQuestionType(job, positionNames, questionType)

	// メッセージの組み立て
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: generateSystemPrompt(job, questionType),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// ChatGPT APIにリクエスト
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    messages,
			Temperature: getTemperatureForQuestionType(questionType), // 質問タイプに応じた柔軟性を設定
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

// 質問タイプに応じたtemperature値を返す
func getTemperatureForQuestionType(questionType string) float32 {
	switch questionType {
	case "skills":
		return 0.5 // スキルは正確さが重要なので低め
	case "learning":
		return 0.7 // 学習方法は少し創造的に
	case "career":
		return 0.8 // キャリアパスはより多様な可能性を示すために高め
	default:
		return 0.6 // 一般的な質問の場合
	}
}

// システムプロンプトを生成する新しい関数
func generateSystemPrompt(job *models.JobPosting, questionType string) string {
	var basePrompt string

	switch questionType {
	case "skills":
		basePrompt = "あなたは技術スキルとキャリア開発のスペシャリストです。ITエンジニアの採用や育成に携わる経験豊富なテクニカルリーダーとして回答してください。"
	case "learning":
		basePrompt = "あなたは技術教育とエンジニア育成の専門家です。効果的な学習方法やリソースに詳しく、段階的なスキル習得プランを提案できるメンターとして回答してください。"
	case "career":
		basePrompt = "あなたはIT業界のキャリアコンサルタントです。技術職のキャリアパスや将来的な成長機会に詳しく、長期的なキャリア設計の視点から助言できるアドバイザーとして回答してください。"
	default:
		basePrompt = "あなたはキャリア開発とスキルアップに関する専門家です。特定の求人に応募するために必要なスキルセットとロードマップを提供してください。"
	}

	// 職種に応じたカスタマイズ
	positionAdaptation := ""
	for _, pos := range job.Positions {
		switch {
		case strings.Contains(strings.ToLower(pos.Name), "フロントエンド"):
			positionAdaptation += "特にフロントエンド開発技術、最新のUIフレームワーク、ユーザー体験の最適化について詳しいです。"
		case strings.Contains(strings.ToLower(pos.Name), "バックエンド"):
			positionAdaptation += "特にバックエンド開発、APIの設計原則、データベース最適化、サーバーインフラについて詳しいです。"
		case strings.Contains(strings.ToLower(pos.Name), "フルスタック"):
			positionAdaptation += "特にフロントエンドとバックエンドの両方の技術スタックに精通し、統合的なシステム開発視点を持っています。"
		case strings.Contains(strings.ToLower(pos.Name), "データ"):
			positionAdaptation += "特にデータ処理、分析技術、ビッグデータアーキテクチャ、データパイプラインの構築について詳しいです。"
		case strings.Contains(strings.ToLower(pos.Name), "ai") || strings.Contains(strings.ToLower(pos.Name), "ml"):
			positionAdaptation += "特に機械学習、AIモデルの設計と学習、自然言語処理、コンピュータビジョンなどの専門知識を持っています。"
		case strings.Contains(strings.ToLower(pos.Name), "devops") || strings.Contains(strings.ToLower(pos.Name), "インフラ"):
			positionAdaptation += "特にCI/CD、クラウドインフラ、コンテナ技術、自動化、システム監視について詳しいです。"
		}
	}

	// レスポンスフォーマットの指示
	formatInstructions := `
回答は以下の形式で構造化してください：
- 見出しを使って情報を整理する
- 重要なポイントには箇条書きを使用
- 具体的な例や推奨事項を含める
- 長すぎず、簡潔で実用的な情報を提供する
- 可能な場合は、時間軸やステップを含める
`

	return basePrompt + " " + positionAdaptation + formatInstructions
}

// 質問タイプに応じたプロンプトを生成する関数
func generatePromptByQuestionType(job *models.JobPosting, positionNames []string, questionType string) string {
	// 共通の求人情報部分
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

	// 求人から主要キーワードを抽出
	keywords := extractKeywords(job)
	keywordStr := ""
	if len(keywords) > 0 {
		keywordStr = "\n\n特に注目すべきキーワード/スキル: " + strings.Join(keywords, ", ")
	}

	// 質問タイプに応じたプロンプトテンプレートを選択
	promptTemplates := getPromptTemplates(questionType)

	// ランダムにテンプレートを選択
	selectedTemplate := promptTemplates[rand.Intn(len(promptTemplates))]

	// テンプレートにジョブ情報を挿入
	specificPrompt := fmt.Sprintf(selectedTemplate, jobInfo+keywordStr)

	return specificPrompt
}

// 質問タイプに応じたプロンプトテンプレートの配列を返す
func getPromptTemplates(questionType string) []string {
	switch questionType {
	case "skills":
		return []string{
			`以下の求人情報に基づいて、この職種に必要なスキルと知識について詳細に解説してください:%s

この回答には以下の要素を含めてください:
1. 必須の技術スキル（プログラミング言語、フレームワーク、ツールなど）の詳細説明
2. あると有利な周辺技術・知識
3. 技術以外の重要なスキル（コミュニケーション能力、問題解決能力など）
4. 各スキルの重要度とスキルレベルの目安

具体的で実践的な説明をお願いします。`,

			`この職種で成功するために必要な技術スキルと非技術スキルを体系的に説明してください:%s

以下の観点から解説をお願いします:
- コア技術スキルとその習熟度レベル
- 差別化できる専門スキル
- 実務で求められる実践的スキル
- 今後重要性が増すと思われる発展的スキル
- チームでの協働に必要な対人スキル

業界の実態に即した現実的な視点からアドバイスをお願いします。`,

			`この求人に応募する場合、採用担当者が最も評価するであろうスキルセットについて解説してください:%s

特に下記の点を含めてください:
- 採用基準を満たす最低限必要なスキル
- 競合他社の応募者と差別化できる特殊スキル
- この職種特有の専門性を示すスキル
- 実務経験を通して証明できるスキル
- 応募書類や面接でアピールすべきスキルのポイント

採用サイドの視点も取り入れた実用的なアドバイスをお願いします。`,
		}

	case "learning":
		return []string{
			`以下の求人情報に基づいて、必要なスキルと知識を効率的に習得するための具体的な学習方法を提案してください:%s

この回答には以下の要素を含めてください:
1. 各技術スキルの効果的な学習リソース（オンラインコース、書籍、チュートリアルなど）
2. 初心者から応募レベルまでの具体的な学習ステップ
3. 推奨の学習プロジェクト
4. 学習の進捗を測定する方法
5. 学習にかかる推定時間

具体的かつ実践的な学習プランを提案してください。`,

			`この職種に必要なスキルを最短で習得するための効率的な学習ロードマップを教えてください:%s

以下の内容を含む学習計画を立ててください:
- スキル習得の優先順位と段階的なアプローチ
- 各スキル向けの具体的なオンラインコースや教材の推奨
- 効果的な学習のためのプロジェクトベースの演習案
- 学習の進捗を確認するための小さなマイルストーン
- より高度なスキルへのステップアップ方法

実務で即戦力となるための実践的なアドバイスをお願いします。`,

			`独学でこの職種に必要なスキルを身につけるための最適な学習方法を教えてください:%s

具体的に以下の点を含めた学習戦略を提案してください:
- 無料と有料の学習リソースのバランスのとれた組み合わせ
- スキルを定着させるための実践的なプロジェクトアイデア
- コミュニティやメンターを見つける方法
- 学習モチベーションを維持するためのテクニック
- 自己学習の効果を最大化するための時間管理術

自己主導型学習の効果を最大化するための実用的なアドバイスをお願いします。`,
		}

	case "career":
		return []string{
			`以下の求人情報に基づいて、この職種のキャリアパスと将来の発展可能性について解説してください:%s

この回答には以下の要素を含めてください:
1. この職種に就いた後のキャリアステップと成長の道筋
2. 次のキャリアレベルに進むために必要なスキルと経験
3. 5年後、10年後のキャリア展望
4. この分野での専門性を深めるための方向性とオプション
5. 関連する他の職種への転向可能性

中長期的なキャリア展望について具体的に解説してください。`,

			`この職種を起点としたキャリア発展の多様な選択肢について詳しく教えてください:%s

以下の観点から将来のキャリアパスについて解説をお願いします:
- 技術専門職としてのキャリアラダー
- マネジメントへの移行ステップ
- 業界内での横展開の可能性
- フリーランスやコンサルタントとしての独立オプション
- 新興技術の発展に伴う将来有望な専門分野

多角的な視点からのキャリア戦略を提示してください。`,

			`この職種で5年以上働いた後のキャリア展望について、実例を交えて教えてください:%s

特に以下の点を盛り込んだ長期的なキャリア戦略を提案してください:
- この業界の成功者に共通するキャリアパターン
- 経験を積んだ後に開ける新たな職種やポジション
- 今後の技術トレンドを見据えた専門性の構築方法
- ワークライフバランスと収入のバランスを考慮したキャリア選択
- 年齢やライフステージに応じたキャリアシフトの選択肢

長期的な視点からのキャリア構築のアドバイスをお願いします。`,
		}

	default: // general
		return []string{
			`以下の求人情報に基づいて、この職種に必要なスキルと知識を習得するためのロードマップを作成してください:%s

このロードマップには以下の要素を含めてください:
1. 必要なプログラミング言語と技術スキルの習得方法
2. おすすめの学習リソース（オンラインコース、書籍、チュートリアルなど）
3. スキルレベル別の学習タイムライン（初心者、中級者、応募準備完了）
4. 必要に応じて英語力向上のための勉強方法

回答は日本語で、具体的かつ実践的なアドバイスをお願いします。`,

			`この職種への転職や応募を検討している方向けの、総合的なアドバイスをお願いします:%s

以下の内容を網羅的に解説してください:
- 求められるスキルセットの明確な説明
- そのスキルを習得するための最適な学習パス
- この職種で成功するための実践的なアドバイス
- 応募書類や面接でアピールすべきポイント
- キャリアの長期的な展望と成長機会

包括的なキャリアガイダンスをお願いします。`,

			`この求人に応募するための準備から面接対策、入社後の成長までを網羅したロードマップを提供してください:%s

特に以下の点について詳しく解説をお願いします:
- 必須スキルと推奨スキルの明確な区別
- 短期間で効率よく必要なスキルを身につける方法
- 応募書類で強調すべき経験やスキル
- 面接でよく聞かれる質問と模範回答
- 入社後のスキルアップ計画

実践的で具体的なステップバイステップのガイドをお願いします。`,
		}
	}
}

// 求人情報から重要なキーワードを抽出する新しい関数
func extractKeywords(job *models.JobPosting) []string {
	// タイトル、説明、要件を結合したテキスト
	combinedText := job.Title + " " + job.Description + " " + job.Requirements
	combinedText = strings.ToLower(combinedText)

	// 技術キーワードのリスト
	techKeywords := []string{
		"java", "python", "go", "golang", "javascript", "typescript",
		"react", "vue", "angular", "node", "express", "nextjs", "nuxt",
		"docker", "kubernetes", "aws", "gcp", "azure", "terraform",
		"sql", "nosql", "mongodb", "mysql", "postgresql", "oracle",
		"graphql", "rest", "api", "microservices", "ci/cd", "git",
		"agile", "scrum", "devops", "machine learning", "ai", "deep learning",
		"tensorflow", "pytorch", "nlp", "computer vision", "data science",
		"big data", "hadoop", "spark", "kafka", "etl", "tableau", "power bi",
		"react native", "flutter", "swift", "kotlin", "ios", "android",
		"html", "css", "sass", "less", "jquery", "bootstrap", "tailwind",
		"php", "laravel", "symfony", "ruby", "rails", "scala", "rust",
		"c#", ".net", "c++", "unity", "game development", "testing", "qa",
		"selenium", "cypress", "jest", "mocha", "linux", "unix", "bash",
		"powershell", "networking", "security", "blockchain", "ethereum",
		"solidity", "smart contracts", "web3", "frontend", "backend", "fullstack",
		"ui/ux", "figma", "sketch", "adobe xd", "photoshop", "illustrator",
	}

	// 見つかったキーワードを保存するマップ
	foundKeywords := make(map[string]bool)

	// すべてのキーワードを検索
	for _, keyword := range techKeywords {
		if strings.Contains(combinedText, keyword) {
			foundKeywords[keyword] = true
		}
	}

	// マップからキーワードリストを作成
	var result []string
	for keyword := range foundKeywords {
		result = append(result, keyword)
	}

	// 結果が多すぎる場合は制限
	if len(result) > 10 {
		result = result[:10]
	}

	return result
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
