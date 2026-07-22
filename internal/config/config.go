package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Security SecurityConfig `yaml:"security"`
	Prompt   PromptConfig   `yaml:"prompt"`
	LLM      LLMConfig      `yaml:"llm"`
}

type ServerConfig struct {
	Port                  int    `yaml:"port"`
	RequestTimeoutSeconds int    `yaml:"request_timeout_seconds"`
	MaxPayloadSizeMB      int    `yaml:"max_payload_size_mb"`
	LogLevel              string `yaml:"log_level"`
}

type SecurityConfig struct {
	APIKey    string          `yaml:"api_key"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

type PromptConfig struct {
	Directory string `yaml:"directory"`
}

type LLMConfig struct {
	Provider              string         `yaml:"provider"`
	DefaultModel          string         `yaml:"default_model"`
	LoadBalancingStrategy string         `yaml:"load_balancing_strategy"`
	OllamaServers         []OllamaServer `yaml:"ollama_servers"`
}

type OllamaServer struct {
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	Weight        int    `yaml:"weight"`
	MaxConcurrent int    `yaml:"max_concurrent"`
}

// LoadConfig reads configuration from a YAML file and overrides with environment variables if set.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:                  8080,
			RequestTimeoutSeconds: 60,
			MaxPayloadSizeMB:      10,
			LogLevel:              "info",
		},
		Security: SecurityConfig{
			APIKey: "krea-secret-ai-key-2026",
			RateLimit: RateLimitConfig{
				RequestsPerMinute: 100,
				Burst:             20,
			},
		},
		Prompt: PromptConfig{
			Directory: "./prompts",
		},
		LLM: LLMConfig{
			Provider:              "ollama",
			DefaultModel:          "qwen3:8b",
			LoadBalancingStrategy: "least_in_flight",
			OllamaServers: []OllamaServer{
				{
					Name:          "ollama-node-1",
					URL:           "http://localhost:11434",
					Weight:        1,
					MaxConcurrent: 10,
				},
			},
		},
	}

	// If path is specified or defaults to config/config.yaml, try reading file; fall back to config/config.example.yaml if config.yaml is missing.
	targetPath := path
	if targetPath == "config/config.yaml" || targetPath == "" {
		if _, err := os.Stat("config/config.yaml"); err == nil {
			targetPath = "config/config.yaml"
		} else if _, err := os.Stat("config/config.example.yaml"); err == nil {
			targetPath = "config/config.example.yaml"
		}
	}

	if targetPath != "" {
		data, err := os.ReadFile(targetPath)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config YAML: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file '%s': %w", targetPath, err)
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = p
		}
	}
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		cfg.Security.APIKey = apiKey
	}
	if promptDir := os.Getenv("PROMPT_DIRECTORY"); promptDir != "" {
		cfg.Prompt.Directory = promptDir
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Server.LogLevel = logLevel
	}
	if model := os.Getenv("MODEL"); model != "" {
		cfg.LLM.DefaultModel = model
	}
	if timeoutStr := os.Getenv("REQUEST_TIMEOUT"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.Server.RequestTimeoutSeconds = t
		}
	}
	// Multi-server Ollama override format: OLLAMA_URL=http://node1:11434,http://node2:11434
	if ollamaURLs := os.Getenv("OLLAMA_URL"); ollamaURLs != "" {
		urls := strings.Split(ollamaURLs, ",")
		var servers []OllamaServer
		for i, url := range urls {
			trimmed := strings.TrimSpace(url)
			if trimmed != "" {
				servers = append(servers, OllamaServer{
					Name:          fmt.Sprintf("ollama-env-%d", i+1),
					URL:           trimmed,
					Weight:        1,
					MaxConcurrent: 10,
				})
			}
		}
		if len(servers) > 0 {
			cfg.LLM.OllamaServers = servers
		}
	}
}
