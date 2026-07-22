package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PromptEngine struct {
	baseDir string
}

func NewEngine(baseDir string) *PromptEngine {
	if baseDir == "" {
		baseDir = "./prompts"
	}
	return &PromptEngine{
		baseDir: baseDir,
	}
}

// Render loads the markdown prompt for the given profile and action, and replaces placeholders with values.
func (e *PromptEngine) Render(profile string, action string, placeholders map[string]string) (string, error) {
	filePath := filepath.Join(e.baseDir, profile, fmt.Sprintf("%s.md", action))

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template file '%s': %w", filePath, err)
	}

	content := string(contentBytes)

	// List of all standard placeholders
	standardKeys := []string{
		"TEXT",
		"TITLE",
		"CONVERSATION",
		"LANGUAGE",
		"TONE",
		"SIGNATURE",
		"CUSTOM_CONTEXT",
	}

	// Apply provided placeholders
	for key, val := range placeholders {
		placeholderTag := fmt.Sprintf("{{%s}}", strings.ToUpper(key))
		content = strings.ReplaceAll(content, placeholderTag, val)
	}

	// Replace any unused standard placeholders with empty string or default fallback
	for _, key := range standardKeys {
		placeholderTag := fmt.Sprintf("{{%s}}", key)
		if strings.Contains(content, placeholderTag) {
			content = strings.ReplaceAll(content, placeholderTag, "")
		}
	}

	return strings.TrimSpace(content), nil
}
