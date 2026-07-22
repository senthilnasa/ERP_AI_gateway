package ticket

import (
	"errors"
	"fmt"
	"strings"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
)

type SupportTicketProfile struct{}

func New() *SupportTicketProfile {
	return &SupportTicketProfile{}
}

func (p *SupportTicketProfile) Name() string {
	return "support_ticket"
}

func (p *SupportTicketProfile) Validate(req *models.WriteRequest) error {
	if strings.TrimSpace(req.Text) == "" {
		return errors.New("support ticket reply text cannot be empty")
	}
	return nil
}

func (p *SupportTicketProfile) BuildPrompt(req *models.WriteRequest, engine *prompt.PromptEngine) (string, error) {
	tone := req.Tone
	if tone == "" {
		tone = "professional and helpful"
	}

	language := req.Language
	if language == "" {
		language = "english"
	}

	var convBuilder strings.Builder
	for _, msg := range req.Context.Conversation {
		convBuilder.WriteString(fmt.Sprintf("%s: %s\n", strings.ToUpper(msg.Role), msg.Message))
	}

	placeholders := map[string]string{
		"TEXT":         req.Text,
		"TITLE":        req.Context.Title,
		"CONVERSATION": convBuilder.String(),
		"TONE":         tone,
		"LANGUAGE":     language,
		"SIGNATURE":    req.Options.Signature,
	}

	return engine.Render(p.Name(), req.Action, placeholders)
}

func (p *SupportTicketProfile) ParseResponse(raw string) (string, error) {
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
