package scripts

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"howtv-server/controllers"
	"howtv-server/models"

	"github.com/google/uuid"
)

type MockCompany struct {
	Name        string           `json:"name"`
	Address     string           `json:"address"`
	Industry    string           `json:"industry"`
	Website     string           `json:"website"`
	JobPostings []MockJobPosting `json:"job_postings"`
}

type MockJobPosting struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Requirements   string   `json:"requirements"`
	SalaryRange    string   `json:"salary_range"`
	Location       string   `json:"location"`
	EmploymentType string   `json:"employment_type"`
	PostingDate    string   `json:"posting_date"`
	ClosingDate    string   `json:"closing_date"`
	Status         string   `json:"status"`
	Positions      []string `json:"positions"`
}

type MockData struct {
	Companies []MockCompany `json:"companies"`
}

// このマップはポジション名を標準化するために使用
var positionNameMap = map[string]string{
	"frontend engineer":       "Frontend Engineer",
	"backend engineer":        "Backend Engineer",
	"fullstack engineer":      "Fullstack Engineer",
	"mobile engineer":         "Mobile Engineer",
	"infrastructure engineer": "Infrastructure Engineer",
	"devops engineer":         "DevOps Engineer",
	"cloud engineer":          "Cloud Engineer",
	"data engineer":           "Data Engineer",
	"ai/ml engineer":          "AI/ML Engineer",
	"security engineer":       "Security Engineer",
	"qa engineer":             "QA Engineer",
	"game engineer":           "Game Engineer",
	"graphics engineer":       "Graphics Engineer",
	"database engineer":       "Database Engineer",
	"solutions architect":     "Solutions Architect",
	"android developer":       "Mobile Engineer",
	"ios developer":           "Mobile Engineer",
	"ruby developer":          "Backend Engineer",
	"python developer":        "Backend Engineer",
}

// ポジションをシードする関数
func SeedPositions() []models.Position {
	standardPositions := []string{
		"Frontend Engineer",
		"Backend Engineer",
		"Fullstack Engineer",
		"Mobile Engineer",
		"Infrastructure Engineer",
		"DevOps Engineer",
		"Cloud Engineer",
		"Data Engineer",
		"AI/ML Engineer",
		"Security Engineer",
		"QA Engineer",
		"Solutions Architect",
		"Game Engineer",
		"Database Engineer",
		"Graphics Engineer",
	}

	log.Printf("ポジションのシード開始: %d件", len(standardPositions))
	createdPositions := make([]models.Position, 0, len(standardPositions))

	for _, posName := range standardPositions {
		var position models.Position
		result := controllers.DB.FirstOrCreate(&position, models.Position{Name: posName})
		if result.Error != nil {
			log.Printf("ポジション作成エラー: %v", result.Error)
			continue
		}
		createdPositions = append(createdPositions, position)
		log.Printf("ポジション作成成功: %s (ID: %d)", position.Name, position.ID)
	}

	// すべてのポジションを取得（既存のものも含む）
	var allPositions []models.Position
	controllers.DB.Find(&allPositions)

	log.Printf("ポジションのシード完了: %d件", len(allPositions))
	return allPositions
}

// 標準化されたポジション名を取得する
func getNormalizedPositionName(name string) string {
	// 小文字に変換して検索
	lowerName := strings.ToLower(name)

	// 直接マッピングがある場合はそれを使用
	if standardName, exists := positionNameMap[lowerName]; exists {
		return standardName
	}

	// Developer/Engineerのマッピング
	lowerName = strings.ReplaceAll(lowerName, "developer", "engineer")

	// もう一度マッピングを確認
	if standardName, exists := positionNameMap[lowerName]; exists {
		return standardName
	}

	// マッピングがなければ最初の文字を大文字にして返す
	words := strings.Fields(lowerName)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[0:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// データベースをシードする関数
func SeedDatabase(filePath string) error {
	// Read mock data file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading mock data file: %v", err)
		return err
	}
	log.Printf("モックデータファイル読み込み成功: %d バイト", len(data))

	// 読み込んだJSONを表示（デバッグ用）
	jsonStr := string(data)
	if len(jsonStr) > 500 {
		log.Printf("モックデータのプレビュー: %s...", jsonStr[:500])
	} else {
		log.Printf("モックデータ全体: %s", jsonStr)
	}

	// Parse mock data
	var mockData MockData
	if err := json.Unmarshal(data, &mockData); err != nil {
		log.Printf("Error parsing mock data: %v", err)
		return err
	}
	log.Printf("モックデータのパース成功: %d社のデータ", len(mockData.Companies))

	// 会社が存在しない場合はエラーを返す
	if len(mockData.Companies) == 0 {
		return fmt.Errorf("モックデータに会社情報がありません")
	}

	// ポジションを先にシード
	positions := SeedPositions()

	// ポジションマップ作成（大文字小文字を区別しない）
	posMap := make(map[string]models.Position)
	for _, pos := range positions {
		normalized := strings.ToLower(pos.Name)
		posMap[normalized] = pos
		log.Printf("ポジションマップに追加: %s -> ID %d", normalized, pos.ID)
	}

	// 求人とポジションの関連付けをカウント
	totalJobs := 0
	totalAssignedPositions := 0

	// 会社データの処理
	for companyIndex, mockCompany := range mockData.Companies {
		log.Printf("処理中の会社 %d/%d: %s", companyIndex+1, len(mockData.Companies), mockCompany.Name)

		// 会社を作成・取得
		company := models.Company{
			Name:     mockCompany.Name,
			Address:  mockCompany.Address,
			Industry: mockCompany.Industry,
			Website:  mockCompany.Website,
		}

		// Create or find company
		result := controllers.DB.FirstOrCreate(&company, models.Company{Name: company.Name})
		if result.Error != nil {
			log.Printf("会社作成エラー %s: %v", company.Name, result.Error)
			continue
		}
		log.Printf("会社作成/取得成功: %s (ID: %d, UUID: %d)", company.Name, company.ID, company.UUID)

		// 会社の求人を処理
		log.Printf("  求人数: %d件", len(mockCompany.JobPostings))
		for jobIndex, mockJob := range mockCompany.JobPostings {
			log.Printf("  処理中の求人 %d/%d: %s", jobIndex+1, len(mockCompany.JobPostings), mockJob.Title)

			// 求人を作成
			job := models.JobPosting{
				UUID:           uuid.New(),
				CompanyID:      company.UUID,
				Title:          mockJob.Title,
				Description:    mockJob.Description,
				Requirements:   mockJob.Requirements,
				SalaryRange:    mockJob.SalaryRange,
				Location:       mockJob.Location,
				EmploymentType: mockJob.EmploymentType,
				Status:         mockJob.Status,
			}

			// 求人を保存
			if err := controllers.DB.Create(&job).Error; err != nil {
				log.Printf("    求人作成エラー: %v", err)
				continue
			}
			totalJobs++
			log.Printf("    求人作成成功: ID: %d, UUID: %s", job.ID, job.UUID)

			// ポジションを割り当て
			if len(mockJob.Positions) > 0 {
				log.Printf("    ポジション指定あり: %v", mockJob.Positions)
				assignPositionsToJob(job, mockJob.Positions, posMap, &totalAssignedPositions)
			} else {
				// ポジション指定がない場合はデフォルトのポジションを割り当て
				log.Printf("    ポジション指定なし: デフォルトポジションを割り当て")
				defaultPositions := []string{"Fullstack Engineer"}
				assignPositionsToJob(job, defaultPositions, posMap, &totalAssignedPositions)
			}
		}
	}

	// データベースの検証を行う
	validateDatabaseSeeding()

	log.Printf("データベースシード完了: %d件の求人に合計%d件のポジションを割り当て", totalJobs, totalAssignedPositions)
	return nil
}

// 求人にポジションを割り当てる関数
func assignPositionsToJob(job models.JobPosting, positionNames []string, posMap map[string]models.Position, totalAssignedPositions *int) {
	if len(positionNames) == 0 {
		log.Printf("      ポジション名が空です")
		return
	}

	// 割り当てるポジションを収集
	var jobPositions []models.Position
	for _, rawPosName := range positionNames {
		if rawPosName == "" {
			continue
		}

		// ポジション名を標準化
		standardName := getNormalizedPositionName(rawPosName)
		normalized := strings.ToLower(standardName)

		// マップからポジションを検索
		if pos, ok := posMap[normalized]; ok {
			jobPositions = append(jobPositions, pos)
			log.Printf("      ポジション '%s' (ID: %d) をマッチ", pos.Name, pos.ID)
		} else {
			// 見つからない場合は新しいポジションを作成
			var newPosition models.Position
			newPosition.Name = standardName

			if err := controllers.DB.FirstOrCreate(&newPosition, models.Position{Name: standardName}).Error; err != nil {
				log.Printf("      新規ポジション作成エラー: %v", err)
				continue
			}

			// マップに追加
			posMap[strings.ToLower(newPosition.Name)] = newPosition
			jobPositions = append(jobPositions, newPosition)
			log.Printf("      新規ポジション '%s' (ID: %d) を作成", newPosition.Name, newPosition.ID)
		}
	}

	// 割り当てるポジションがある場合
	if len(jobPositions) > 0 {
		// トランザクションを開始
		tx := controllers.DB.Begin()

		// 一度クリア
		if err := tx.Model(&job).Association("Positions").Clear(); err != nil {
			tx.Rollback()
			log.Printf("      ポジション関連クリアエラー: %v", err)
			return
		}

		// ポジションを関連付け
		for _, pos := range jobPositions {
			// 中間テーブルに直接挿入（ORMのバグ回避のため）
			if err := tx.Exec("INSERT INTO job_positions (job_posting_id, position_id) VALUES (?, ?)", job.ID, pos.ID).Error; err != nil {
				log.Printf("      ポジション割り当てエラー (SQL): %v", err)
				continue
			}
			*totalAssignedPositions++
			log.Printf("      ポジション割り当て成功: Job ID %d -> Position ID %d", job.ID, pos.ID)
		}

		// コミット
		if err := tx.Commit().Error; err != nil {
			log.Printf("      トランザクションコミットエラー: %v", err)
			return
		}

		log.Printf("      求人 '%s' に %d 件のポジションを割り当て成功", job.Title, len(jobPositions))
	} else {
		log.Printf("      求人 '%s' には割り当て可能なポジションがありませんでした", job.Title)
	}
}

// データベースのポジション関連をリセットする関数（デバッグ・修正用）
func ResetPositionAssociations() error {
	// 中間テーブルをクリア
	if err := controllers.DB.Exec("DELETE FROM job_positions").Error; err != nil {
		return fmt.Errorf("ポジション関連のリセットに失敗しました: %w", err)
	}

	log.Printf("job_positionsテーブルをクリアしました")

	// すべての求人を取得
	var jobs []models.JobPosting
	if err := controllers.DB.Find(&jobs).Error; err != nil {
		return fmt.Errorf("求人取得エラー: %w", err)
	}

	log.Printf("%d件の求人のポジション関連をリセットしました", len(jobs))
	return nil
}

// データベースのシード状態を検証する関数
func validateDatabaseSeeding() {
	var (
		positionCount int64
		jobCount      int64
		relationCount int64
	)

	controllers.DB.Model(&models.Position{}).Count(&positionCount)
	controllers.DB.Model(&models.JobPosting{}).Count(&jobCount)
	controllers.DB.Table("job_positions").Count(&relationCount)

	log.Printf("データベース検証: ポジション %d件, 求人 %d件, 関連 %d件", positionCount, jobCount, relationCount)

	// ポジションが割り当てられていない求人を検出
	var jobsWithoutPositions []models.JobPosting
	controllers.DB.Joins("LEFT JOIN job_positions ON job_postings.id = job_positions.job_posting_id").
		Where("job_positions.position_id IS NULL").
		Distinct().
		Find(&jobsWithoutPositions)

	if len(jobsWithoutPositions) > 0 {
		log.Printf("警告: %d件の求人にポジションが割り当てられていません", len(jobsWithoutPositions))

		// 未割り当て求人にデフォルトポジションを割り当て
		var defaultPosition models.Position
		if err := controllers.DB.FirstOrCreate(&defaultPosition, models.Position{Name: "Fullstack Engineer"}).Error; err != nil {
			log.Printf("デフォルトポジション取得エラー: %v", err)
			return
		}

		for _, job := range jobsWithoutPositions {
			// 中間テーブルに直接挿入
			err := controllers.DB.Exec("INSERT INTO job_positions (job_posting_id, position_id) VALUES (?, ?)",
				job.ID, defaultPosition.ID).Error
			if err != nil {
				log.Printf("ポジション自動割り当てエラー: %v", err)
			} else {
				log.Printf("求人 ID %d にデフォルトポジション '%s' を割り当てました", job.ID, defaultPosition.Name)
			}
		}
	}

	// 各ポジションの求人数を表示
	log.Println("各ポジションの求人数:")
	var positions []models.Position
	controllers.DB.Find(&positions)

	for _, pos := range positions {
		var count int64
		controllers.DB.Table("job_positions").Where("position_id = ?", pos.ID).Count(&count)
		log.Printf("  %s (ID: %d): %d件", pos.Name, pos.ID, count)
	}
}
