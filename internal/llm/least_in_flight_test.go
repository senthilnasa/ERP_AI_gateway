package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/senthilnasa/ERP_AI_gateway/internal/config"
)

func TestLeastInFlightLoadBalancer(t *testing.T) {
	// Create 2 mock Ollama HTTP servers with artificial delay to verify least-in-flight distribution
	var server1Count, server2Count int64
	var mu sync.Mutex

	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		server1Count++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OllamaGenerateResponse{
			Model:    "qwen3:8b",
			Response: "Response from Server 1",
			Done:     true,
		})
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		server2Count++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OllamaGenerateResponse{
			Model:    "qwen3:8b",
			Response: "Response from Server 2",
			Done:     true,
		})
	}))
	defer srv2.Close()

	llmCfg := config.LLMConfig{
		Provider:              "ollama",
		DefaultModel:          "qwen3:8b",
		LoadBalancingStrategy: "least_in_flight",
		OllamaServers: []config.OllamaServer{
			{Name: "Node-1", URL: srv1.URL, Weight: 1, MaxConcurrent: 10},
			{Name: "Node-2", URL: srv2.URL, Weight: 1, MaxConcurrent: 10},
		},
	}

	provider := NewMultiServerOllamaProvider(llmCfg, 5)

	// Launch 10 concurrent requests
	var wg sync.WaitGroup
	numRequests := 10
	errChan := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := provider.Generate(context.Background(), "Hello", "")
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Fatalf("unexpected error during generation: %v", err)
	}

	mu.Lock()
	c1 := server1Count
	c2 := server2Count
	mu.Unlock()

	t.Logf("Server 1 requests: %d, Server 2 requests: %d", c1, c2)

	if c1 == 0 || c2 == 0 {
		t.Errorf("expected load balancing across both servers, got Server1=%d, Server2=%d", c1, c2)
	}

	stats := provider.GetBackendStats()
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
}
