package profile

import "testing"

func TestCleanLLMOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Python code block inside conversational wrapper with bold quotes",
			input: "Understood! I'll respond in the following format:\n\n" +
				"```python\n" +
				"\"**This is test ticket and I'm closing this**\"\n" +
				"```\n\n" +
				"Please let me know if you'd like any modifications or additional content added to make it more coherent!",
			expected: "This is test ticket and I'm closing this",
		},
		{
			name: "Triple quotes enclosure",
			input: "Sure, here is the text:\n\n" +
				"\"\"\"This is a test ticket response\"\"\"\n\n" +
				"Hope this helps!",
			expected: "This is a test ticket response",
		},
		{
			name: "Single backtick wrapping inline text",
			input: "`This is a test ticket response`",
			expected: "This is a test ticket response",
		},
		{
			name: "Code block without specified language",
			input: "```\nIssue has been resolved and verified.\n```",
			expected: "Issue has been resolved and verified.",
		},
		{
			name: "Intro Response: header",
			input: "Response:\nDear Customer, your ticket #1024 has been resolved.",
			expected: "Dear Customer, your ticket #1024 has been resolved.",
		},
		{
			name: "Outro with Feel free to ask",
			input: "Hello Team, the build is ready.\n\nFeel free to ask if you have any questions.",
			expected: "Hello Team, the build is ready.",
		},
		{
			name: "Plain multiline text without wrappers",
			input: "Dear User,\n\nWe have updated your account balance.\n\nRegards,\nFinance Team",
			expected: "Dear User,\n\nWe have updated your account balance.\n\nRegards,\nFinance Team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanLLMOutput(tt.input)
			if got != tt.expected {
				t.Errorf("CleanLLMOutput() = %q, want %q", got, tt.expected)
			}
		})
	}
}
