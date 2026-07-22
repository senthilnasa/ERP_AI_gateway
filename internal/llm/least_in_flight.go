package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/senthilnasa/ERP_AI_gateway/internal/config"
	"github.com/senthilnasa/ERP_AI_gateway/internal/logger"
)

type OllamaNode struct {
	Name          string
	URL           string
	Weight        int
	MaxConcurrent int

	inFlight      int64
	totalRequests int64
	errorCount    int64

	mu        sync.RWMutex
	isHealthy bool
}

type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaGenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

type MultiServerOllamaProvider struct {
	nodes        []*OllamaNode
	httpClient   *http.Client
	defaultModel string
	strategy     string
	mu           sync.Mutex
	rrIndex      uint64
}

func NewMultiServerOllamaProvider(cfg config.LLMConfig, timeoutSec int) *MultiServerOllamaProvider {
	nodes := make([]*OllamaNode, 0, len(cfg.OllamaServers))
	for _, s := range cfg.OllamaServers {
		nodes = append(nodes, &OllamaNode{
			Name:          s.Name,
			URL:           s.URL,
			Weight:        s.Weight,
			MaxConcurrent: s.MaxConcurrent,
			isHealthy:     true,
		})
	}

	return &MultiServerOllamaProvider{
		nodes: nodes,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		defaultModel: cfg.DefaultModel,
		strategy:     cfg.LoadBalancingStrategy,
	}
}

func (p *MultiServerOllamaProvider) Name() string {
	return "ollama-least-in-flight"
}

func (p *MultiServerOllamaProvider) DefaultModel() string {
	return p.defaultModel
}

func (p *MultiServerOllamaProvider) GetBackendStats() []BackendStat {
	stats := make([]BackendStat, len(p.nodes))
	for i, node := range p.nodes {
		node.mu.RLock()
		healthy := node.isHealthy
		node.mu.RUnlock()

		stats[i] = BackendStat{
			Name:          node.Name,
			URL:           node.URL,
			InFlight:      atomic.LoadInt64(&node.inFlight),
			TotalRequests: atomic.LoadInt64(&node.totalRequests),
			ErrorCount:    atomic.LoadInt64(&node.errorCount),
			IsHealthy:     healthy,
		}
	}
	return stats
}

// selectNodeLeastInFlight selects the healthy node with the smallest number of in-flight requests.
func (p *MultiServerOllamaProvider) selectNodeLeastInFlight() *OllamaNode {
	var selected *OllamaNode
	minInFlight := int64(^uint64(0) >> 1) // max int64

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, node := range p.nodes {
		node.mu.RLock()
		healthy := node.isHealthy
		node.mu.RUnlock()

		if !healthy {
			continue
		}

		currentInFlight := atomic.LoadInt64(&node.inFlight)
		if currentInFlight < minInFlight {
			minInFlight = currentInFlight
			selected = node
		}
	}

	// Fallback to round-robin if no healthy node was found with least-in-flight check
	if selected == nil && len(p.nodes) > 0 {
		idx := atomic.AddUint64(&p.rrIndex, 1)
		selected = p.nodes[idx%uint64(len(p.nodes))]
	}

	return selected
}

func (p *MultiServerOllamaProvider) Generate(ctx context.Context, prompt string, model string) (string, error) {
	if model == "" {
		model = p.defaultModel
	}

	if len(p.nodes) == 0 {
		return "", fmt.Errorf("no Ollama backend servers configured")
	}

	var lastErr error
	triedNodes := make(map[string]bool)

	// Try up to len(nodes) times in case of backend network failure
	for attempt := 0; attempt < len(p.nodes); attempt++ {
		node := p.selectNodeLeastInFlight()
		if node == nil || triedNodes[node.URL] {
			break
		}
		triedNodes[node.URL] = true

		atomic.AddInt64(&node.inFlight, 1)
		atomic.AddInt64(&node.totalRequests, 1)

		resp, err := p.executeOnNode(ctx, node, prompt, model)
		atomic.AddInt64(&node.inFlight, -1)

		if err == nil {
			// Successful execution reset error count & ensure healthy state
			node.mu.Lock()
			node.isHealthy = true
			node.mu.Unlock()
			return resp, nil
		}

		// Log failure and update health metrics
		logger.Get().Warn("Ollama node %s (%s) request failed: %v", node.Name, node.URL, err)
		errCount := atomic.AddInt64(&node.errorCount, 1)
		if errCount >= 3 {
			node.mu.Lock()
			node.isHealthy = false
			node.mu.Unlock()
		}

		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("all Ollama backends failed, last error: %w", lastErr)
	}

	return "", fmt.Errorf("failed to process completion request across Ollama cluster")
}

func (p *MultiServerOllamaProvider) executeOnNode(ctx context.Context, node *OllamaNode, prompt string, model string) (string, error) {
	reqBody := OllamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/generate", node.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create http request for %s: %w", node.Name, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed on %s: %w", node.Name, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from %s: %w", node.Name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama backend %s returned status %d: %s", node.Name, resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp OllamaGenerateResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response from %s: %w", node.Name, err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error from %s: %s", node.Name, ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}
