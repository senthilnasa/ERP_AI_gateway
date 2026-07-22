package profile

import (
	"regexp"
	"strings"
)

var (
	// Regex matching code blocks ```[lang]\n ... ```
	codeBlockRegex = regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_-]+)?\\s*\n?(.*?)\n?```")

	// Regex matching intro phrases
	introPhrases = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(?:understoood|understood|sure|certainly|here is|here's|here are|here's a|here is a|below is|i'd be happy to|response:|draft:)[^\n]*\n+`),
	}

	// Regex matching outro phrases
	outroPhrases = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\n+(?:please let me know|let me know|hope this helps|feel free to|is there anything else|let me know if you'd like)[^\n]*$`),
	}
)

// CleanLLMOutput strips code fences, conversational intro/outro lines, and enclosing quotes/formatting from raw LLM output.
func CleanLLMOutput(raw string) string {
	cleaned := strings.TrimSpace(raw)

	// 1. If code block exists anywhere in the text, extract content inside code block
	matches := codeBlockRegex.FindStringSubmatch(cleaned)
	if len(matches) > 1 {
		cleaned = strings.TrimSpace(matches[1])
	} else {
		// Strip isolated ``` lines if present
		cleaned = regexp.MustCompile("(?m)^```[a-zA-Z0-9_-]*$").ReplaceAllString(cleaned, "")
		cleaned = strings.TrimSpace(cleaned)
	}

	// 2. Remove conversational intro lines
	for _, re := range introPhrases {
		cleaned = re.ReplaceAllString(cleaned, "")
	}
	cleaned = strings.TrimSpace(cleaned)

	// 3. Remove conversational outro lines
	for _, re := range outroPhrases {
		cleaned = re.ReplaceAllString(cleaned, "")
	}
	cleaned = strings.TrimSpace(cleaned)

	// 4. Strip surrounding quotes ("...") or bold (**...**) if the entire text is wrapped in them
	if (strings.HasPrefix(cleaned, `"`) && strings.HasSuffix(cleaned, `"`)) ||
		(strings.HasPrefix(cleaned, `'`) && strings.HasSuffix(cleaned, `'`)) {
		if len(cleaned) >= 2 {
			cleaned = strings.TrimSpace(cleaned[1 : len(cleaned)-1])
		}
	}

	if strings.HasPrefix(cleaned, "**") && strings.HasSuffix(cleaned, "**") && len(cleaned) >= 4 {
		cleaned = strings.TrimSpace(cleaned[2 : len(cleaned)-2])
	}

	return strings.TrimSpace(cleaned)
}
