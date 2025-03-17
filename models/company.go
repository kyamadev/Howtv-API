package models

import (
	"gorm.io/gorm"
)

type Company struct {
	gorm.Model
	UUID        uint         `json:"uuid" gorm:"primarykey"`
	Name        string       `json:"name"`
	Address     string       `json:"address"`
	Industry    string       `json:"industry"`
	Website     string       `json:"website"`
	LogoURL     string       `json:"logo_url"`
	JobPostings []JobPosting `json:"job_postings" gorm:"foreignKey:CompanyID"`
}
