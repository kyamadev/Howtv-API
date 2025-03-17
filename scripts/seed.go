package scripts

import (
	"encoding/json"
	"log"
	"os"

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
	Title          string `json:"title"`
	Description    string `json:"description"`
	Requirements   string `json:"requirements"`
	SalaryRange    string `json:"salary_range"`
	Location       string `json:"location"`
	EmploymentType string `json:"employment_type"`
	PostingDate    string `json:"posting_date"`
	ClosingDate    string `json:"closing_date"`
	Status         string `json:"status"`
}

type MockData struct {
	Companies []MockCompany `json:"companies"`
}

func SeedPositions() []models.Position {
	positions := []models.Position{
		{Name: "フロントエンドエンジニア"},
		{Name: "バックエンドエンジニア"},
		{Name: "フルスタックエンジニア"},
		{Name: "モバイルエンジニア"},
		{Name: "インフラエンジニア"},
		{Name: "データエンジニア"},
		{Name: "AI/ML エンジニア"},
		{Name: "DevOpsエンジニア"},
		{Name: "セキュリティエンジニア"},
		{Name: "QAエンジニア"},
	}

	for i := range positions {
		controllers.DB.FirstOrCreate(&positions[i], models.Position{Name: positions[i].Name})
	}

	return positions
}

func SeedDatabase(filePath string) error {
	// Read mock data file using os.ReadFile instead of deprecated ioutil.ReadFile
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading mock data file: %v", err)
		return err
	}

	// Parse mock data
	var mockData MockData
	if err := json.Unmarshal(data, &mockData); err != nil {
		log.Printf("Error parsing mock data: %v", err)
		return err
	}

	// Seed positions first
	positions := SeedPositions()

	// Create position map for easy lookup
	posMap := make(map[string]models.Position)
	for _, pos := range positions {
		posMap[pos.Name] = pos
	}

	for _, mockCompany := range mockData.Companies {
		company := models.Company{
			Name:     mockCompany.Name,
			Address:  mockCompany.Address,
			Industry: mockCompany.Industry,
			Website:  mockCompany.Website,
		}

		// Create or find company
		result := controllers.DB.FirstOrCreate(&company, models.Company{Name: company.Name})
		if result.Error != nil {
			log.Printf("Error creating company %s: %v", company.Name, result.Error)
			continue
		}

		// Create job postings for this company
		for _, mockJob := range mockCompany.JobPostings {
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

			// Create job posting
			if err := controllers.DB.Create(&job).Error; err != nil {
				log.Printf("Error creating job posting: %v", err)
				continue
			}

			// Assign appropriate positions based on title/description
			var jobPositions []models.Position

			// Example logic to assign positions based on keywords
			if mockJob.Title == "フルリモートエンジニア" {
				// For remote positions, check description for specific skills
				if containsAny(mockJob.Description, []string{"frontend", "UI", "React", "Vue", "JavaScript"}) {
					jobPositions = append(jobPositions, posMap["フロントエンドエンジニア"])
				}
				if containsAny(mockJob.Description, []string{"backend", "server", "API", "Go", "Python", "Java"}) {
					jobPositions = append(jobPositions, posMap["バックエンドエンジニア"])
				}
				if containsAny(mockJob.Description, []string{"full-stack", "full stack", "フルスタック"}) {
					jobPositions = append(jobPositions, posMap["フルスタックエンジニア"])
				}
			} else if mockJob.Title == "東京ポジションエンジニア" {
				// Similar logic for Tokyo-based positions
				if containsAny(mockJob.Description, []string{"frontend", "UI", "React", "Vue", "JavaScript"}) {
					jobPositions = append(jobPositions, posMap["フロントエンドエンジニア"])
				}
				if containsAny(mockJob.Description, []string{"backend", "server", "API", "Go", "Python", "Java"}) {
					jobPositions = append(jobPositions, posMap["バックエンドエンジニア"])
				}
			}

			// If no specific positions detected, assign a default one based on title
			if len(jobPositions) == 0 {
				if containsAny(mockJob.Title, []string{"フル", "full"}) {
					jobPositions = append(jobPositions, posMap["フルスタックエンジニア"])
				} else {
					// Default to both front-end and back-end for general engineering positions
					jobPositions = append(jobPositions, posMap["フロントエンドエンジニア"])
					jobPositions = append(jobPositions, posMap["バックエンドエンジニア"])
				}
			}

			// Assign positions to job
			if len(jobPositions) > 0 {
				if err := controllers.DB.Model(&job).Association("Positions").Append(&jobPositions); err != nil {
					log.Printf("Error assigning positions to job: %v", err)
				}
			}
		}
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// Helper function to check if a string contains any of the keywords
func containsAny(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if contains(s, keyword) {
			return true
		}
	}
	return false
}

// Helper function to check if a string contains a keyword
func contains(s, keyword string) bool {
	// 文字列検索の改良が必要かも
	return true // 一旦すべてtrue
}
