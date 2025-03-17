package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"howtv-server/config"
	"howtv-server/models"
	"howtv-server/services"
)

func GenerateRoadmap(c *gin.Context) {
	id := c.Param("uuid")

	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なUUID形式です"})
		return
	}

	var job models.JobPosting
	if err := DB.Preload("Positions").Where("uuid = ?", jobUUID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "求人情報が見つかりませんでした"})
		return
	}

	apiKey := config.GetOpenAIAPIKey()
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI APIキーが設定されていません"})
		return
	}

	openaiService := services.NewOpenAIService(apiKey)

	roadmap, err := openaiService.GenerateCareerRoadmap(&job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_title": job.Title,
		"location":  job.Location,
		"roadmap":   roadmap.Roadmap,
	})
}
