package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig("non_existent_config.yaml")
	if err != nil {
		t.Fatalf("unexpected error loading missing config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.LLM.LoadBalancingStrategy != "least_in_flight" {
		t.Errorf("expected strategy least_in_flight, got %s", cfg.LLM.LoadBalancingStrategy)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("OLLAMA_URL", "http://node1:11434,http://node2:11434")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("OLLAMA_URL")
	}()

	cfg, err := LoadConfig("non_existent.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if len(cfg.LLM.OllamaServers) != 2 {
		t.Fatalf("expected 2 ollama servers from env, got %d", len(cfg.LLM.OllamaServers))
	}
	if cfg.LLM.OllamaServers[0].URL != "http://node1:11434" {
		t.Errorf("expected node1 URL, got %s", cfg.LLM.OllamaServers[0].URL)
	}
	if cfg.LLM.OllamaServers[1].URL != "http://node2:11434" {
		t.Errorf("expected node2 URL, got %s", cfg.LLM.OllamaServers[1].URL)
	}
}
