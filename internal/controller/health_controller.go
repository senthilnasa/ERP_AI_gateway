package controller

import (
	"net/http"

	"github.com/senthilnasa/ERP_AI_gateway/internal/service"
	"github.com/gin-gonic/gin"
)

type HealthController struct {
	aiService    *service.AIService
	defaultModel string
}

func NewHealthController(aiService *service.AIService, defaultModel string) *HealthController {
	return &HealthController{
		aiService:    aiService,
		defaultModel: defaultModel,
	}
}

func (ctrl *HealthController) HealthCheck(c *gin.Context) {
	stats := ctrl.aiService.GetBackendStats()
	c.JSON(http.StatusOK, gin.H{
		"status":          "UP",
		"service":         "OneERP AI Gateway",
		"ollama_backends": stats,
	})
}

func (ctrl *HealthController) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "1.0.0",
		"build":   "2026-07-22",
		"engine":  "Go 1.24+ / Gin",
	})
}

func (ctrl *HealthController) ListProfiles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"profiles": ctrl.aiService.GetProfiles(),
	})
}

func (ctrl *HealthController) ListActions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"actions": ctrl.aiService.GetActions(),
	})
}

func (ctrl *HealthController) ListModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"default_model": ctrl.defaultModel,
		"available_models": []string{
			"qwen3:8b",
			"qwen3:14b",
			"llama3:8b",
			"mistral:7b",
		},
		"backend_nodes": ctrl.aiService.GetBackendStats(),
	})
}
