package controllers

import (
	"net/http"

	"howtv-server/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetJobPostings returns all job postings with their positions
func GetJobPostings(c *gin.Context) {
	var jobs []models.JobPosting
	if err := DB.Preload("Positions").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetJobPosting returns a single job posting with its positions
func GetJobPosting(c *gin.Context) {
	id := c.Param("uuid")

	// Parse UUID
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var job models.JobPosting
	if err := DB.Preload("Positions").Where("uuid = ?", jobUUID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job posting not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// CreateJobPosting creates a new job posting
func CreateJobPosting(c *gin.Context) {
	var jobDTO struct {
		models.JobPosting
		PositionIDs []uint `json:"position_ids"`
	}

	if err := c.ShouldBindJSON(&jobDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a new UUID if not provided
	if jobDTO.JobPosting.UUID == uuid.Nil {
		jobDTO.JobPosting.UUID = uuid.New()
	}

	// Start a transaction
	tx := DB.Begin()

	// Create the job posting
	if err := tx.Create(&jobDTO.JobPosting).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Assign positions if any are provided
	if len(jobDTO.PositionIDs) > 0 {
		var positions []models.Position
		if err := tx.Where("id IN ?", jobDTO.PositionIDs).Find(&positions).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
			return
		}

		if err := tx.Model(&jobDTO.JobPosting).Association("Positions").Append(positions); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign positions"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return the created job with positions
	var createdJob models.JobPosting
	if err := DB.Preload("Positions").Where("uuid = ?", jobDTO.JobPosting.UUID).First(&createdJob).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job created but failed to retrieve it"})
		return
	}

	c.JSON(http.StatusCreated, createdJob)
}

// UpdateJobPosting updates a job posting
func UpdateJobPosting(c *gin.Context) {
	id := c.Param("uuid")

	// Parse UUID
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var jobDTO struct {
		models.JobPosting
		PositionIDs []uint `json:"position_ids"`
	}

	if err := c.ShouldBindJSON(&jobDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if job exists
	var job models.JobPosting
	if err := DB.Where("uuid = ?", jobUUID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job posting not found"})
		return
	}

	// Start a transaction
	tx := DB.Begin()

	// Update job posting
	jobDTO.JobPosting.UUID = jobUUID
	if err := tx.Model(&models.JobPosting{}).Where("uuid = ?", jobUUID).Updates(jobDTO.JobPosting).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update positions if provided
	if len(jobDTO.PositionIDs) > 0 {
		// Clear existing positions
		if err := tx.Model(&job).Association("Positions").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear positions"})
			return
		}

		// Assign new positions
		var positions []models.Position
		if err := tx.Where("id IN ?", jobDTO.PositionIDs).Find(&positions).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
			return
		}

		if err := tx.Model(&job).Association("Positions").Append(positions); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign positions"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return the updated job with positions
	var updatedJob models.JobPosting
	if err := DB.Preload("Positions").Where("uuid = ?", jobUUID).First(&updatedJob).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job updated but failed to retrieve it"})
		return
	}

	c.JSON(http.StatusOK, updatedJob)
}

// DeleteJobPosting deletes a job posting
func DeleteJobPosting(c *gin.Context) {
	id := c.Param("uuid")

	// Parse UUID
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := DB.Where("uuid = ?", jobUUID).Delete(&models.JobPosting{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job posting deleted successfully"})
}
