package models

import "gorm.io/gorm"

type JobPosting struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey" json:"id"`
	CompanyID      uint   `json:"company_id"`
	Position       string `json:"position"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Requirements   string `json:"requirements"`
	SalaryRange    string `json:"salary_range"`
	Location       string `json:"location"`
	EmploymentType string `json:"employment_type"`
	// PostingDate    time.Time `json:"posting_date"`
	// ClosingDate    time.Time `json:"closing_date"`
	Status string `json:"status"`
}
