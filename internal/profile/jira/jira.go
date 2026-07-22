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

// ParseJiraStorySections parses a generated Jira story string into title, description, and acceptance criteria sections.
func ParseJiraStorySections(raw string) (title, description, ac string) {
	lines := strings.Split(raw, "\n")
	var currentSection string
	var titleLines, descLines, acLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "title:") {
			t := strings.TrimSpace(line[6:])
			if t != "" {
				titleLines = append(titleLines, t)
			}
			currentSection = "title"
			continue
		} else if strings.HasPrefix(lower, "description:") {
			rest := strings.TrimSpace(line[12:])
			if rest != "" {
				descLines = append(descLines, rest)
			}
			currentSection = "desc"
			continue
		} else if strings.HasPrefix(lower, "acceptance criteria:") || strings.HasPrefix(lower, "acceptance criteria") {
			var rest string
			if strings.HasPrefix(lower, "acceptance criteria:") {
				rest = strings.TrimSpace(line[20:])
			} else {
				rest = strings.TrimSpace(line[19:])
			}
			if rest != "" {
				acLines = append(acLines, rest)
			}
			currentSection = "ac"
			continue
		}

		switch currentSection {
		case "title":
			if trimmed != "" && len(titleLines) == 0 {
				titleLines = append(titleLines, trimmed)
			}
		case "desc":
			descLines = append(descLines, line)
		case "ac":
			acLines = append(acLines, line)
		}
	}

	title = strings.TrimSpace(strings.Join(titleLines, " "))
	description = strings.TrimSpace(strings.Join(descLines, "\n"))
	ac = strings.TrimSpace(strings.Join(acLines, "\n"))
	return
}
