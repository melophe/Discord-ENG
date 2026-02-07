package bot

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// onMessageCreate handles incoming messages (for reply-based answers)
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	// Check if this is a reply to a quiz message
	if m.ReferencedMessage == nil {
		return
	}

	// Check if the referenced message is from our bot
	if m.ReferencedMessage.Author.ID != s.State.User.ID {
		return
	}

	// Check if it's a quiz message by looking for the embed
	if len(m.ReferencedMessage.Embeds) == 0 {
		return
	}

	embed := m.ReferencedMessage.Embeds[0]
	if embed.Title == "" || len(embed.Title) < 10 {
		return
	}

	// Extract question ID from title (e.g., "📝 英作文問題 #42")
	questionID := extractQuestionID(embed.Title)
	if questionID == 0 {
		return
	}

	// Get the question from database
	question, err := b.db.GetQuestion(questionID)
	if err != nil {
		log.Printf("Error getting question: %v", err)
		return
	}

	// Show typing indicator
	s.ChannelTyping(m.ChannelID)

	// Evaluate the answer using Claude
	ctx := context.Background()
	result, err := b.claude.EvaluateAnswer(ctx, question.Japanese, m.Content)
	if err != nil {
		log.Printf("Error evaluating answer: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ 回答の評価に失敗しました")
		return
	}

	// Save the answer to database
	err = b.db.SaveAnswer(m.Author.ID, questionID, m.Content, result.ModelAnswer, result.Score, result.Feedback)
	if err != nil {
		log.Printf("Error saving answer: %v", err)
	}

	// Create response embed
	responseEmbed := b.createEvaluationEmbed(m.Content, result.Score, result.Feedback, result.ModelAnswer)

	// Send response
	_, err = s.ChannelMessageSendEmbed(m.ChannelID, responseEmbed)
	if err != nil {
		log.Printf("Error sending evaluation: %v", err)
	}
}

// extractQuestionID extracts question ID from title like "📝 英作文問題 #42"
func extractQuestionID(title string) int64 {
	re := regexp.MustCompile(`#(\d+)`)
	matches := re.FindStringSubmatch(title)
	if len(matches) < 2 {
		return 0
	}
	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// createEvaluationEmbed creates the evaluation response embed
func (b *Bot) createEvaluationEmbed(userAnswer string, score int, feedback, modelAnswer string) *discordgo.MessageEmbed {
	// Choose color based on score
	var color int
	var emoji string
	switch {
	case score >= 90:
		color = 0x00D4AA // Green
		emoji = "🎉"
	case score >= 70:
		color = 0x5865F2 // Blue
		emoji = "👍"
	case score >= 50:
		color = 0xFFA500 // Orange
		emoji = "📝"
	default:
		color = 0xFF6B6B // Red
		emoji = "💪"
	}

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s 回答評価", emoji),
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "あなたの回答",
				Value: userAnswer,
			},
			{
				Name:   "📊 スコア",
				Value:  fmt.Sprintf("**%d** / 100", score),
				Inline: true,
			},
			{
				Name:  "📖 模範解答",
				Value: modelAnswer,
			},
			{
				Name:  "💬 フィードバック",
				Value: feedback,
			},
		},
	}
}
