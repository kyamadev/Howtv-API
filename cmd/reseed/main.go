package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"howtv-server/controllers"
	"howtv-server/models"
	"howtv-server/scripts"
)

func main() {
	// コマンドライン引数からDBパスを取得
	dbPath := "test.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	fmt.Printf("データベースのリセットと再シードを行います: %s\n", dbPath)

	// データベース接続
	var err error
	controllers.DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}

	// 現在のデータ状況を確認
	var positionCount, jobCount, associationCount int64
	controllers.DB.Model(&models.Position{}).Count(&positionCount)
	controllers.DB.Model(&models.JobPosting{}).Count(&jobCount)
	controllers.DB.Table("job_positions").Count(&associationCount)

	fmt.Printf("現在のデータ状態: ポジション %d件, 求人 %d件, 関連 %d件\n",
		positionCount, jobCount, associationCount)

	// 既存のポジション関連をリセット
	fmt.Println("既存のポジション関連をリセットします...")
	if err := scripts.ResetPositionAssociations(); err != nil {
		fmt.Printf("警告: ポジション関連のリセットに失敗しました: %v\n", err)
	}

	// データベースのシード
	mockDataPath := "mockdata.txt"
	fmt.Printf("データベースをシードします（データファイル: %s）...\n", mockDataPath)

	if err := scripts.SeedDatabase(mockDataPath); err != nil {
		fmt.Printf("警告: データベースのシードに失敗しました: %v\n", err)
	} else {
		fmt.Println("データベースのシードに成功しました")
	}

	// シード後のデータ確認
	controllers.DB.Model(&models.Position{}).Count(&positionCount)
	controllers.DB.Model(&models.JobPosting{}).Count(&jobCount)
	controllers.DB.Table("job_positions").Count(&associationCount)

	fmt.Printf("シード後のデータ状態: ポジション %d件, 求人 %d件, 関連 %d件\n",
		positionCount, jobCount, associationCount)

	// 各ポジションの求人数を確認
	var positions []models.Position
	controllers.DB.Find(&positions)

	fmt.Println("\n各ポジションの求人数:")
	fmt.Println("---------------------------")

	for _, pos := range positions {
		var count int64
		controllers.DB.Table("job_positions").Where("position_id = ?", pos.ID).Count(&count)
		fmt.Printf("%-25s: %d件\n", pos.Name, count)
	}

	fmt.Println("\nデータベースの再シードが完了しました")
}
