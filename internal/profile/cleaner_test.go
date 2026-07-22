package profile

import "testing"

func TestCleanLLMOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Code block inside conversational wrapper",
			input: "Understood! I'll respond in the following format:\n\n" +
				"```python\n" +
				"\"**This is test ticket and I'm closing this**\"\n" +
				"```\n\n" +
				"Please let me know if you'd like any modifications or additional content added to make it more coherent!",
			expected: "This is test ticket and I'm closing this",
		},
		{
			name: "Plain text with intro and outro",
			input: "Here is the revised message:\n\n" +
				"Dear Customer, your ticket has been resolved.\n\n" +
				"Hope this helps!",
			expected: "Dear Customer, your ticket has been resolved.",
		},
		{
			name:     "Quoted text inside code block",
			input:    "```\n\"Hello World\"\n```",
			expected: "Hello World",
		},
		{
			name:     "Clean response",
			input:    "This is a clean response.",
			expected: "This is a clean response.",
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
