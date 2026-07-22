package llm

import (
	"context"
	"fmt"
	"sync/atomic"
)

type MockLLMProvider struct {
	requestsCount int64
}

func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{}
}

func (m *MockLLMProvider) Name() string {
	return "mock-provider"
}

func (m *MockLLMProvider) DefaultModel() string {
	return "mock:model"
}

func (m *MockLLMProvider) GetBackendStats() []BackendStat {
	count := atomic.LoadInt64(&m.requestsCount)
	return []BackendStat{
		{
			Name:          "mock-backend",
			URL:           "mock://internal",
			InFlight:      0,
			TotalRequests: count,
			ErrorCount:    0,
			IsHealthy:     true,
		},
	}
}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string, model string) (string, error) {
	atomic.AddInt64(&m.requestsCount, 1)
	return fmt.Sprintf("[Mock Output - Model: %s] Successfully processed request. Prompt length: %d chars.", model, len(prompt)), nil
}
