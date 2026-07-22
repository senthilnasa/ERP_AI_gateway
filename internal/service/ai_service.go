package service

import (
	"context"
	"fmt"
	"time"

	"github.com/senthilnasa/ERP_AI_gateway/internal/action"
	"github.com/senthilnasa/ERP_AI_gateway/internal/llm"
	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
)

type AIService struct {
	profiles *profile.Registry
	actions  *action.Registry
	prompts  *prompt.PromptEngine
	llm      llm.LLM
}

func NewAIService(profiles *profile.Registry, actions *action.Registry, prompts *prompt.PromptEngine, llm llm.LLM) *AIService {
	return &AIService{
		profiles: profiles,
		actions:  actions,
		prompts:  prompts,
		llm:      llm,
	}
}

func (s *AIService) ProcessWrite(ctx context.Context, req *models.WriteRequest) (*models.WriteResponseData, error) {
	start := time.Now()

	// 1. Resolve Profile
	p, err := s.profiles.Get(req.Profile)
	if err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	// 2. Resolve Action
	_, err = s.actions.Get(req.Action)
	if err != nil {
		return nil, fmt.Errorf("invalid action: %w", err)
	}

	// 3. Validate Profile specific input
	if err := p.Validate(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 4. Render Prompt using Prompt Engine
	promptStr, err := p.BuildPrompt(req, s.prompts)
	if err != nil {
		return nil, fmt.Errorf("prompt rendering error: %w", err)
	}

	// 5. Model Selection (Request options override -> Default model)
	targetModel := req.Options.Model
	if targetModel == "" {
		targetModel = s.llm.DefaultModel()
	}

	// 6. Invoke LLM Provider (Routed via Least-in-flight multi-server load balancer)
	rawLLMOutput, err := s.llm.Generate(ctx, promptStr, targetModel)
	if err != nil {
		return nil, fmt.Errorf("llm generation error: %w", err)
	}

	// 7. Parse and clean response via Profile
	finalResult, err := p.ParseResponse(rawLLMOutput)
	if err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	elapsedMs := time.Since(start).Milliseconds()

	return &models.WriteResponseData{
		Result:       finalResult,
		Profile:      req.Profile,
		Action:       req.Action,
		Model:        targetModel,
		ProcessingMS: elapsedMs,
	}, nil
}

func (s *AIService) GetProfiles() []string {
	return s.profiles.ListNames()
}

func (s *AIService) GetActions() []string {
	return s.actions.ListNames()
}

func (s *AIService) GetBackendStats() []llm.BackendStat {
	return s.llm.GetBackendStats()
}
