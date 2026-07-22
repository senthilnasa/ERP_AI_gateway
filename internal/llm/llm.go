package llm

import (
	"context"
)

type BackendStat struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	InFlight      int64  `json:"in_flight"`
	TotalRequests int64  `json:"total_requests"`
	ErrorCount    int64  `json:"error_count"`
	IsHealthy     bool   `json:"is_healthy"`
}

type LLM interface {
	Generate(ctx context.Context, prompt string, model string) (string, error)
	Name() string
	DefaultModel() string
	GetBackendStats() []BackendStat
}
