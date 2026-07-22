package inline

import (
	"errors"
	"strings"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
)

type InlineProfile struct{}

func New() *InlineProfile {
	return &InlineProfile{}
}

func (p *InlineProfile) Name() string {
	return "inline_text"
}

func (p *InlineProfile) Validate(req *models.WriteRequest) error {
	if strings.TrimSpace(req.Text) == "" {
		return errors.New("inline text field cannot be empty")
	}
	return nil
}

func (p *InlineProfile) BuildPrompt(req *models.WriteRequest, engine *prompt.PromptEngine) (string, error) {
	tone := req.Tone
	if tone == "" {
		tone = "professional"
	}

	language := req.Language
	if language == "" {
		language = "english"
	}

	placeholders := map[string]string{
		"TEXT":     req.Text,
		"TONE":     tone,
		"LANGUAGE": language,
	}

	return engine.Render(p.Name(), req.Action, placeholders)
}

func (p *InlineProfile) ParseResponse(raw string) (string, error) {
	return profile.CleanLLMOutput(raw), nil
}
