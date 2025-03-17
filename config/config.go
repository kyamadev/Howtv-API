package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIAPIKey string
}

var (
	instance *Config
	once     sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		instance = &Config{}

		envPaths := []string{
			".env",
			"../.env",
			"../../.env",
			filepath.Join(getProjectRoot(), ".env"), // プロジェクトルート
		}

		loaded := false
		for _, path := range envPaths {
			if _, err := os.Stat(path); err == nil {
				if err := godotenv.Load(path); err == nil {
					log.Printf("環境変数を読み込みました: %s", path)
					loaded = true
					break
				}
			}
		}

		if !loaded {
			log.Println(".envファイルが見つかりませんでした。環境変数から直接読み込みます。")
		}

		instance.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")

		// 設定の検証とログ出力
		validateAndLogConfig()
	})

	return instance
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filepath.Dir(filename))

	return dir
}

func validateAndLogConfig() {
	if instance.OpenAIAPIKey == "" {
		log.Println("警告: OPENAI_API_KEY が設定されていません")
	} else {
		maskedKey := maskAPIKey(instance.OpenAIAPIKey)
		log.Printf("OpenAI APIキーが設定されています: %s", maskedKey)
	}
}

// APIキーの一部をマスクして表示
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func GetOpenAIAPIKey() string {
	if instance == nil {
		LoadConfig()
	}

	return instance.OpenAIAPIKey
}
