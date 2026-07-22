package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/senthilnasa/ERP_AI_gateway/internal/action"
	"github.com/senthilnasa/ERP_AI_gateway/internal/llm"
	"github.com/senthilnasa/ERP_AI_gateway/internal/middleware"
	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/email"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/inline"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/jira"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/ticket"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
	"github.com/senthilnasa/ERP_AI_gateway/internal/service"
	"github.com/gin-gonic/gin"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create prompt directory and files
	promptsDir := t.TempDir()
	emailDir := promptsDir + "/email"
	os.MkdirAll(emailDir, 0755)
	os.WriteFile(emailDir+"/rewrite.md", []byte("Rewrite: {{TEXT}}"), 0644)

	jiraDir := promptsDir + "/jira_story"
	os.MkdirAll(jiraDir, 0755)
	os.WriteFile(jiraDir+"/generate.md", []byte("Generate Jira Story: {{TEXT}}"), 0644)

	promptEngine := prompt.NewEngine(promptsDir)

	profileReg := profile.NewRegistry()
	profileReg.Register(email.New())
	profileReg.Register(ticket.New())
	profileReg.Register(inline.New())
	profileReg.Register(jira.New())

	actionReg := action.NewRegistry()

	mockLLM := llm.NewMockLLMProvider()
	aiSvc := service.NewAIService(profileReg, actionReg, promptEngine, mockLLM)

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.AuthMiddleware("test-api-key"))

	writeCtrl := NewWriteController(aiSvc)
	healthCtrl := NewHealthController(aiSvc, "qwen3:8b")
	futureCtrl := NewFutureStubsController()
	swaggerCtrl := NewSwaggerController("../../docs/swagger.json")

	r.GET("/favicon.ico", HandleFavicon)
	r.GET("/health", healthCtrl.HealthCheck)
	r.GET("/version", healthCtrl.Version)
	r.GET("/profiles", healthCtrl.ListProfiles)
	r.GET("/actions", healthCtrl.ListActions)
	r.GET("/models", healthCtrl.ListModels)
	r.GET("/docs", swaggerCtrl.ServeUI)
	r.GET("/docs/swagger.json", swaggerCtrl.ServeJSON)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/write", writeCtrl.HandleWrite)
		v1.POST("/speech/transcribe", futureCtrl.TranscribeSpeech)
	}

	r.NoRoute(HandleNotFound)

	return r
}

func TestHealthEndpoints(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /health, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/profiles", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /profiles, got %d", w.Code)
	}
}

func TestSwaggerEndpoints(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/docs", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /docs, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/docs/swagger.json", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /docs/swagger.json, got %d", w.Code)
	}
}

func TestNotFoundHandler(t *testing.T) {
	router := setupTestRouter(t)

	// HTML 404 for root or web pages
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for /, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "404 - Page Not Found") {
		t.Errorf("expected Generic 404 HTML, got %s", w.Body.String())
	}

	// JSON 404 for /api/* paths
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/invalid_path", nil)
	req.Header.Set("Authorization", "Bearer test-api-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for /api/v1/invalid_path, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "NOT_FOUND") {
		t.Errorf("expected JSON NOT_FOUND code, got %s", w.Body.String())
	}
}

func TestFaviconHandler(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/favicon.ico", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /favicon.ico, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "image/svg+xml" {
		t.Errorf("expected image/svg+xml Content-Type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestWriteAPIUnauthorized(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	reqBody, _ := json.Marshal(models.WriteRequest{
		Profile: "email",
		Action:  "rewrite",
		Text:    "test email text",
	})
	req, _ := http.NewRequest("POST", "/api/v1/write", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", w.Code)
	}
}

func TestWriteAPISuccess(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	reqBody, _ := json.Marshal(models.WriteRequest{
		Profile: "email",
		Action:  "rewrite",
		Text:    "please process my order",
		Tone:    "professional",
	})
	req, _ := http.NewRequest("POST", "/api/v1/write", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer test-api-key")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
}

func TestFutureStubsNotImplemented(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/speech/transcribe", nil)
	req.Header.Set("Authorization", "Bearer test-api-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented, got %d", w.Code)
	}
}

func TestJiraStoryGeneration(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	reqBody, _ := json.Marshal(models.WriteRequest{
		Profile: "jira_story",
		Action:  "generate",
		Text:    "Allow users to reset their password via SMS OTP",
	})
	req, _ := http.NewRequest("POST", "/api/v1/write", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer test-api-key")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for jira_story, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
}
