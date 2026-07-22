package controller

import (
	"fmt"
	"net/http"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/service"
	"github.com/gin-gonic/gin"
)

type WriteController struct {
	aiService *service.AIService
}

func NewWriteController(aiService *service.AIService) *WriteController {
	return &WriteController{
		aiService: aiService,
	}
}

func (ctrl *WriteController) HandleWrite(c *gin.Context) {
	reqID, _ := c.Get("RequestID")
	reqIDStr := fmt.Sprintf("%v", reqID)

	var req models.WriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error: models.ApiErrorDetail{
				Code:      "BAD_REQUEST",
				Message:   fmt.Sprintf("Invalid request payload: %v", err),
				RequestID: reqIDStr,
			},
		})
		return
	}

	if req.Metadata.RequestID == "" {
		req.Metadata.RequestID = reqIDStr
	}

	resData, err := ctrl.aiService.ProcessWrite(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, models.ApiResponse{
			Success: false,
			Error: models.ApiErrorDetail{
				Code:      "PROCESSING_FAILED",
				Message:   err.Error(),
				RequestID: reqIDStr,
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    resData,
	})
}
