package email

import (
	"errors"
	"strings"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/profile"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
)

type EmailProfile struct{}

func New() *EmailProfile {
	return &EmailProfile{}
}

func (p *EmailProfile) Name() string {
	return "email"
}

func (p *EmailProfile) Validate(req *models.WriteRequest) error {
	if strings.TrimSpace(req.Text) == "" {
		return errors.New("email text field cannot be empty")
	}
	return nil
}

func (p *EmailProfile) BuildPrompt(req *models.WriteRequest, engine *prompt.PromptEngine) (string, error) {
	tone := req.Tone
	if tone == "" {
		tone = "professional"
	}

	language := req.Language
	if language == "" {
		language = "english"
	}

	placeholders := map[string]string{
		"TEXT":           req.Text,
		"TONE":           tone,
		"LANGUAGE":       language,
		"SIGNATURE":      req.Options.Signature,
		"CUSTOM_CONTEXT": req.Context.Title,
	}

	return engine.Render(p.Name(), req.Action, placeholders)
}

func (p *EmailProfile) ParseResponse(raw string) (string, error) {
	return profile.CleanLLMOutput(raw), nil
}
