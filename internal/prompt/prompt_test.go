package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptEngineRender(t *testing.T) {
	tempDir := t.TempDir()
	emailDir := filepath.Join(tempDir, "email")
	os.MkdirAll(emailDir, 0755)

	templateContent := `Tone: {{TONE}}
Language: {{LANGUAGE}}
Body: {{TEXT}}
Signature: {{SIGNATURE}}`

	os.WriteFile(filepath.Join(emailDir, "rewrite.md"), []byte(templateContent), 0644)

	engine := NewEngine(tempDir)
	res, err := engine.Render("email", "rewrite", map[string]string{
		"TONE":      "Professional",
		"LANGUAGE":  "English",
		"TEXT":      "Please review document.",
		"SIGNATURE": "Support Team",
	})

	if err != nil {
		t.Fatalf("failed to render prompt: %v", err)
	}

	expected := `Tone: Professional
Language: English
Body: Please review document.
Signature: Support Team`

	if res != expected {
		t.Errorf("rendered prompt mismatch.\nExpected:\n%s\nGot:\n%s", expected, res)
	}
}
