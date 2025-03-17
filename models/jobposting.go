package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobPosting struct {
	gorm.Model
	UUID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"uuid"`
	CompanyID      uint       `json:"company_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Requirements   string     `json:"requirements"`
	SalaryRange    string     `json:"salary_range"`
	Location       string     `json:"location"`
	EmploymentType string     `json:"employment_type"`
	Status         string     `json:"status"`
	Positions      []Position `gorm:"many2many:job_positions" json:"positions"`
}

// UUID生成
func (jp *JobPosting) BeforeCreate(tx *gorm.DB) error {
	if jp.UUID == uuid.Nil {
		jp.UUID = uuid.New()
	}
	return nil
}
