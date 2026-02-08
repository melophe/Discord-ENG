package claude

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Client struct {
	client *anthropic.Client
	model  anthropic.Model
}

// NewClient creates a new Claude API client
func NewClient(apiKey, model string) *Client {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		client: &client,
		model:  anthropic.Model(model),
	}
}

// GenerateQuestion generates a Japanese sentence for English translation practice
func (c *Client) GenerateQuestion(ctx context.Context, theme, difficulty string) (string, error) {
	prompt := fmt.Sprintf(GenerateQuestionPrompt, theme, difficulty)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 200,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate question: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return message.Content[0].Text, nil
}

// EvaluationResult holds the result of answer evaluation
type EvaluationResult struct {
	Score       int    `json:"score"`
	Feedback    string `json:"feedback"`
	ModelAnswer string `json:"model_answer"`
}

// EvaluateAnswer evaluates the user's English translation
func (c *Client) EvaluateAnswer(ctx context.Context, japanese, userAnswer string) (*EvaluationResult, error) {
	prompt := fmt.Sprintf(EvaluateAnswerPrompt, japanese, userAnswer)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 500,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answer: %w", err)
	}

	if len(message.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse the response
	result := parseEvaluationResponse(message.Content[0].Text)
	return result, nil
}

// parseEvaluationResponse parses Claude's evaluation response
func parseEvaluationResponse(response string) *EvaluationResult {
	result := &EvaluationResult{
		Score:       70,
		Feedback:    response,
		ModelAnswer: "",
	}

	// Simple parsing - extract score, model answer, and feedback
	lines := splitLines(response)
	for i, line := range lines {
		if len(line) > 7 && line[:6] == "SCORE:" {
			fmt.Sscanf(line, "SCORE: %d", &result.Score)
		} else if len(line) > 14 && line[:13] == "MODEL_ANSWER:" {
			result.ModelAnswer = line[14:]
		} else if len(line) > 10 && line[:9] == "FEEDBACK:" {
			// Collect remaining lines as feedback
			result.Feedback = line[10:]
			for j := i + 1; j < len(lines); j++ {
				result.Feedback += "\n" + lines[j]
			}
			break
		}
	}

	return result
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	var current string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
