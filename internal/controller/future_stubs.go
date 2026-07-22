package controller

import (
	"fmt"
	"net/http"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/gin-gonic/gin"
)

type FutureStubsController struct{}

func NewFutureStubsController() *FutureStubsController {
	return &FutureStubsController{}
}

func (f *FutureStubsController) TranscribeSpeech(c *gin.Context) {
	reqID, _ := c.Get("RequestID")
	c.JSON(http.StatusNotImplemented, models.ApiResponse{
		Success: false,
		Error: models.ApiErrorDetail{
			Code:      "NOT_IMPLEMENTED",
			Message:   "Speech Transcribe (Whisper) endpoint is scheduled for Future Features phase",
			RequestID: fmt.Sprintf("%v", reqID),
		},
	})
}

func (f *FutureStubsController) SynthesizeSpeech(c *gin.Context) {
	reqID, _ := c.Get("RequestID")
	c.JSON(http.StatusNotImplemented, models.ApiResponse{
		Success: false,
		Error: models.ApiErrorDetail{
			Code:      "NOT_IMPLEMENTED",
			Message:   "Speech Synthesize (Piper) endpoint is scheduled for Future Features phase",
			RequestID: fmt.Sprintf("%v", reqID),
		},
	})
}

func (f *FutureStubsController) SummarizeDocument(c *gin.Context) {
	reqID, _ := c.Get("RequestID")
	c.JSON(http.StatusNotImplemented, models.ApiResponse{
		Success: false,
		Error: models.ApiErrorDetail{
			Code:      "NOT_IMPLEMENTED",
			Message:   "Document Summarize endpoint is scheduled for Future Features phase",
			RequestID: fmt.Sprintf("%v", reqID),
		},
	})
}

func (f *FutureStubsController) AnalyzeImage(c *gin.Context) {
	reqID, _ := c.Get("RequestID")
	c.JSON(http.StatusNotImplemented, models.ApiResponse{
		Success: false,
		Error: models.ApiErrorDetail{
			Code:      "NOT_IMPLEMENTED",
			Message:   "Image OCR & Understanding endpoint is scheduled for Future Features phase",
			RequestID: fmt.Sprintf("%v", reqID),
		},
	})
}

func (f *FutureStubsController) QueryRAG(c *gin.Context) {
	reqID, _ := c.Get("RequestID")
	c.JSON(http.StatusNotImplemented, models.ApiResponse{
		Success: false,
		Error: models.ApiErrorDetail{
			Code:      "NOT_IMPLEMENTED",
			Message:   "RAG Knowledge Base Query (Qdrant) endpoint is scheduled for Future Features phase",
			RequestID: fmt.Sprintf("%v", reqID),
		},
	})
}
