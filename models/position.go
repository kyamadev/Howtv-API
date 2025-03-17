package models

import "gorm.io/gorm"

type Position struct {
	gorm.Model
	ID   uint         `gorm:"primaryKey" json:"id"`
	Name string       `json:"name" gorm:"uniqueIndex"`
	Jobs []JobPosting `gorm:"many2many:job_positions" json:"jobs,omitempty"`
}
