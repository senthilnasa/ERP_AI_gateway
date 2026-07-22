package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/senthilnasa/ERP_AI_gateway/internal/action"
	"github.com/senthilnasa/ERP_AI_gateway/internal/config"
	"github.com/senthilnasa/ERP_AI_gateway/internal/controller"
	"github.com/senthilnasa/ERP_AI_gateway/internal/llm"
	"github.com/senthilnasa/ERP_AI_gateway/internal/logger"
	"github.com/senthilnasa/ERP_AI_gateway/internal/middleware"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/email"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/inline"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/jira"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile/ticket"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
	"github.com/senthilnasa/ERP_AI_gateway/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.Init(cfg.Server.LogLevel)
	log.Info("Starting OneERP AI Gateway v1.0.0...")
	log.Info("Loaded configuration from '%s'", configPath)

	// 1. LLM Provider Setup with Multi-Server Load Balancer
	var llmProvider llm.LLM
	if cfg.LLM.Provider == "mock" {
		log.Info("Using Mock LLM Provider")
		llmProvider = llm.NewMockLLMProvider()
	} else {
		log.Info("Using Multi-Server Ollama Provider with '%s' load balancing strategy", cfg.LLM.LoadBalancingStrategy)
		for _, node := range cfg.LLM.OllamaServers {
			log.Info(" - Registered Backend Node: %s -> %s (max_concurrent: %d)", node.Name, node.URL, node.MaxConcurrent)
		}
		llmProvider = llm.NewMultiServerOllamaProvider(cfg.LLM, cfg.Server.RequestTimeoutSeconds)
	}

	// 2. Prompt Engine Setup
	promptEngine := prompt.NewEngine(cfg.Prompt.Directory)

	// 3. Profile Registry Setup
	profileRegistry := profile.NewRegistry()
	profileRegistry.Register(email.New())
	profileRegistry.Register(ticket.New())
	profileRegistry.Register(inline.New())
	profileRegistry.Register(jira.New())
	log.Info("Registered Profiles: %v", profileRegistry.ListNames())

	// 4. Action Registry Setup
	actionRegistry := action.NewRegistry()
	log.Info("Registered Actions: %v", actionRegistry.ListNames())

	// 5. AI Service Core Setup
	aiService := service.NewAIService(profileRegistry, actionRegistry, promptEngine, llmProvider)

	// 6. Router & Middleware Setup
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.RateLimitMiddleware(cfg.Security.RateLimit))
	router.Use(middleware.AuthMiddleware(cfg.Security.APIKey))
	router.Use(middleware.MaxBytesMiddleware(cfg.Server.MaxPayloadSizeMB))

	// 7. Controller Wiring
	writeCtrl := controller.NewWriteController(aiService)
	healthCtrl := controller.NewHealthController(aiService, cfg.LLM.DefaultModel)
	futureCtrl := controller.NewFutureStubsController()
	swaggerCtrl := controller.NewSwaggerController("docs/swagger.json")

	// Public / Operational Endpoints
	router.GET("/favicon.ico", controller.HandleFavicon)
	router.GET("/health", healthCtrl.HealthCheck)
	router.GET("/version", healthCtrl.Version)
	router.GET("/profiles", healthCtrl.ListProfiles)
	router.GET("/actions", healthCtrl.ListActions)
	router.GET("/models", healthCtrl.ListModels)

	// Interactive Swagger API Documentation UI
	router.GET("/docs", swaggerCtrl.ServeUI)
	router.GET("/docs/swagger.json", swaggerCtrl.ServeJSON)

	// Core API v1
	v1 := router.Group("/api/v1")
	{
		v1.POST("/write", writeCtrl.HandleWrite)

		// Future Feature API Stubs
		v1.POST("/speech/transcribe", futureCtrl.TranscribeSpeech)
		v1.POST("/speech/synthesize", futureCtrl.SynthesizeSpeech)
		v1.POST("/document/summarize", futureCtrl.SummarizeDocument)
		v1.POST("/image/analyze", futureCtrl.AnalyzeImage)
		v1.POST("/rag/query", futureCtrl.QueryRAG)
	}

	// 404 Not Found Custom Handler (HTML for web / JSON for /api/*)
	router.NoRoute(controller.HandleNotFound)

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second,
	}

	go func() {
		log.Info("HTTP Server listening on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down AI Gateway server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped cleanly.")
}
