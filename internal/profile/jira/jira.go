package jira

import (
	"errors"
	"strings"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
)

type JiraStoryProfile struct{}

func New() *JiraStoryProfile {
	return &JiraStoryProfile{}
}

func (p *JiraStoryProfile) Name() string {
	return "jira_story"
}

func (p *JiraStoryProfile) Validate(req *models.WriteRequest) error {
	if strings.TrimSpace(req.Text) == "" {
		return errors.New("jira_story request text cannot be empty")
	}
	return nil
}

func (p *JiraStoryProfile) BuildPrompt(req *models.WriteRequest, engine *prompt.PromptEngine) (string, error) {
	tone := req.Tone
	if tone == "" {
		tone = "professional and structured"
	}

	language := req.Language
	if language == "" {
		language = "english"
	}

	placeholders := map[string]string{
		"TEXT":     req.Text,
		"TITLE":    req.Context.Title,
		"TONE":     tone,
		"LANGUAGE": language,
	}

	action := req.Action
	if action == "" {
		action = "generate"
	}

	return engine.Render(p.Name(), action, placeholders)
}

func (p *JiraStoryProfile) ParseResponse(raw string) (string, error) {
	cleaned := strings.TrimSpace(raw)
	if strings.HasPrefix(cleaned, "```") {
		lines := strings.Split(cleaned, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[0], "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[:len(lines)-1]
			}
			cleaned = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
	return cleaned, nil
}
