package controllers

import (
	"net/http"

	"howtv-server/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetPositions(c *gin.Context) {
	var positions []models.Position
	if err := DB.Find(&positions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, positions)
}

func CreatePosition(c *gin.Context) {
	var position models.Position
	if err := c.ShouldBindJSON(&position); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Create(&position).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, position)
}

func AssignPositionsToJob(c *gin.Context) {
	id := c.Param("uuid")

	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var job models.JobPosting
	if err := DB.Where("uuid = ?", jobUUID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job posting not found"})
		return
	}

	var positions []int
	if err := c.ShouldBindJSON(&positions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Model(&job).Association("Positions").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, posID := range positions {
		var pos models.Position
		if err := DB.First(&pos, posID).Error; err != nil {
			continue
		}
		if err := DB.Model(&job).Association("Positions").Append(&pos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Positions assigned successfully"})
}
