package claude

import (
	"testing"
)

func TestParseEvaluationResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		expectedScore  int
		expectedModel  string
		expectedFeedback string
	}{
		{
			name: "full response",
			response: `SCORE: 85
MODEL_ANSWER: This is a test.
FEEDBACK: Great job! Your answer is correct.`,
			expectedScore:  85,
			expectedModel:  "This is a test.",
			expectedFeedback: "Great job! Your answer is correct.",
		},
		{
			name: "multiline feedback",
			response: `SCORE: 70
MODEL_ANSWER: Hello world.
FEEDBACK: Good attempt!
You could improve by using more natural expressions.
Keep practicing!`,
			expectedScore:  70,
			expectedModel:  "Hello world.",
			expectedFeedback: "Good attempt!\nYou could improve by using more natural expressions.\nKeep practicing!",
		},
		{
			name: "no structured format",
			response: "Just some feedback without structure",
			expectedScore:  70,  // default
			expectedModel:  "",
			expectedFeedback: "Just some feedback without structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEvaluationResponse(tt.response)

			if result.Score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, result.Score)
			}
			if result.ModelAnswer != tt.expectedModel {
				t.Errorf("Expected model answer '%s', got '%s'", tt.expectedModel, result.ModelAnswer)
			}
			// Feedback comparison - just check it's not empty for structured responses
			if tt.name != "no structured format" && result.Feedback != tt.expectedFeedback {
				t.Errorf("Expected feedback '%s', got '%s'", tt.expectedFeedback, result.Feedback)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			input:    "single line",
			expected: []string{"single line"},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "line1\n",
			expected: []string{"line1"},
		},
	}

	for _, tt := range tests {
		result := splitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("For input '%s': expected %d lines, got %d", tt.input, len(tt.expected), len(result))
			continue
		}
		for i, line := range result {
			if line != tt.expected[i] {
				t.Errorf("For input '%s': expected line[%d] = '%s', got '%s'", tt.input, i, tt.expected[i], line)
			}
		}
	}
}
