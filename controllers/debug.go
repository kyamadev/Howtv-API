package controllers

import (
	"net/http"
	"os"

	"howtv-server/models"

	"github.com/gin-gonic/gin"
)

// DataStats は現在のデータベース状態に関する統計を返す
func DataStats(c *gin.Context) {
	// ポジション数をカウント
	var positionCount int64
	DB.Model(&models.Position{}).Count(&positionCount)

	// 求人数をカウント
	var jobCount int64
	DB.Model(&models.JobPosting{}).Count(&jobCount)

	// 会社数をカウント
	var companyCount int64
	DB.Model(&models.Company{}).Count(&companyCount)

	// 中間テーブルのレコード数をカウント
	var relationCount int64
	DB.Table("job_positions").Count(&relationCount)

	// 各ポジションの求人数を取得
	type PositionStat struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}

	var positions []models.Position
	DB.Find(&positions)

	positionStats := make([]PositionStat, 0, len(positions))
	for _, pos := range positions {
		var count int64
		DB.Table("job_positions").Where("position_id = ?", pos.ID).Count(&count)
		positionStats = append(positionStats, PositionStat{
			ID:    pos.ID,
			Name:  pos.Name,
			Count: count,
		})
	}

	// 添付ファイルのないジョブを検出
	var jobsWithoutPositions []models.JobPosting
	DB.Joins("LEFT JOIN job_positions ON job_postings.id = job_positions.job_posting_id").
		Where("job_positions.position_id IS NULL").
		Distinct().
		Find(&jobsWithoutPositions)

	c.JSON(http.StatusOK, gin.H{
		"positions_count":        positionCount,
		"jobs_count":             jobCount,
		"companies_count":        companyCount,
		"relations_count":        relationCount,
		"position_stats":         positionStats,
		"jobs_without_positions": len(jobsWithoutPositions),
	})
}

// FixMissingPositions は未割り当ての求人にデフォルトのポジションを割り当てる
func FixMissingPositions(c *gin.Context) {
	// ポジションが割り当てられていない求人を見つける
	var jobsWithoutPositions []models.JobPosting

	DB.Joins("LEFT JOIN job_positions ON job_postings.id = job_positions.job_posting_id").
		Where("job_positions.position_id IS NULL").
		Distinct().
		Find(&jobsWithoutPositions)

	if len(jobsWithoutPositions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "すべての求人にポジションが割り当てられています",
			"fixed":   0,
		})
		return
	}

	// デフォルトポジションを取得
	var defaultPosition models.Position
	result := DB.FirstOrCreate(&defaultPosition, models.Position{Name: "Fullstack Engineer"})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "デフォルトポジションの取得に失敗しました"})
		return
	}

	// 修正カウンター
	fixedCount := 0

	// 未割り当て求人にデフォルトポジションを割り当て
	for _, job := range jobsWithoutPositions {
		if err := DB.Model(&job).Association("Positions").Append(&defaultPosition); err != nil {
			continue
		}
		fixedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                      "未割り当ての求人にデフォルトポジションを割り当てました",
		"total_jobs_without_positions": len(jobsWithoutPositions),
		"fixed":                        fixedCount,
	})
}

// ResetAndReseed はデータベースをリセットして再シードする
func ResetAndReseed(c *gin.Context) {
	// このAPIはセキュリティのため開発環境でのみ有効にすべき
	if gin.Mode() != gin.DebugMode {
		c.JSON(http.StatusForbidden, gin.H{"error": "This endpoint is only available in debug mode"})
		return
	}

	// トランザクション開始
	tx := DB.Begin()

	// 中間テーブルをクリア
	if err := tx.Exec("DELETE FROM job_positions").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "中間テーブルのクリアに失敗しました"})
		return
	}

	// 求人をクリア
	if err := tx.Exec("DELETE FROM job_postings").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "求人テーブルのクリアに失敗しました"})
		return
	}

	// 会社をクリア
	if err := tx.Exec("DELETE FROM companies").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "会社テーブルのクリアに失敗しました"})
		return
	}

	// ポジションをクリア
	if err := tx.Exec("DELETE FROM positions").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ポジションテーブルのクリアに失敗しました"})
		return
	}

	// トランザクションコミット
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "トランザクションのコミットに失敗しました"})
		return
	}

	// シードスクリプトを実行
	// ここではimportの関係で直接実行できないため、環境変数を設定して
	// サーバーの次回起動時にシードされるようにフラグを立てる
	err := os.Setenv("RESEED", "true")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "環境変数の設定に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "データベースをクリアしました。サーバーを再起動して再シードしてください。",
	})
}
