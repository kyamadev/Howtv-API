package models

import (
	"gorm.io/gorm"
)

type company struct {
	gorm.Model
	UUID        uint         `json:"uuid" gorm:"primarykey"`
	Name        string       `json:"name"`
	Addewss     string       `json:"address"`
	Website     string       `json:"website"`
	LogoURL     string       `json:"logo_url"`
	JobPostings []JobPosting `json:"job_postings gorm:"foreignKey:CompanyID"`
}
