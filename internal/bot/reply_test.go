package bot

import (
	"testing"
)

func TestExtractQuestionID(t *testing.T) {
	tests := []struct {
		title    string
		expected int64
	}{
		{"📝 英作文問題 #1", 1},
		{"📝 英作文問題 #42", 42},
		{"📝 英作文問題 #123456", 123456},
		{"Random title", 0},
		{"No number here", 0},
		{"#abc not a number", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result := extractQuestionID(tt.title)
		if result != tt.expected {
			t.Errorf("For title '%s': expected %d, got %d", tt.title, tt.expected, result)
		}
	}
}
