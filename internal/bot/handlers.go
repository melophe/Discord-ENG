package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// onInteractionCreate handles slash commands and button interactions
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleSlashCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponentInteraction(s, i)
	case discordgo.InteractionModalSubmit:
		b.handleModalSubmit(s, i)
	}
}

// handleSlashCommand handles slash command interactions
func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "quiz":
		b.handleQuizCommand(s, i)
	case "theme":
		b.handleThemeCommand(s, i)
	case "stats":
		b.handleStatsCommand(s, i)
	case "settings":
		b.handleSettingsCommand(s, i)
	}
}

// handleQuizCommand generates and sends a new quiz question
func (b *Bot) handleQuizCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer response to avoid timeout
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	userID := i.Member.User.ID
	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		b.respondError(s, i, "エラーが発生しました")
		return
	}

	// Generate question using Claude
	ctx := context.Background()
	japanese, err := b.claude.GenerateQuestion(ctx, user.Theme, user.Difficulty)
	if err != nil {
		log.Printf("Error generating question: %v", err)
		b.respondError(s, i, "問題の生成に失敗しました")
		return
	}

	// Save question to database
	questionID, err := b.db.SaveQuestion(japanese, user.Difficulty, user.Theme)
	if err != nil {
		log.Printf("Error saving question: %v", err)
	}

	// Create quiz message with buttons
	embed := b.createQuizEmbed(questionID, japanese, user.Theme, user.Difficulty)
	components := b.createQuizButtons()

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	if err != nil {
		log.Printf("Error sending quiz: %v", err)
	}
}

// handleThemeCommand sets the user's quiz theme
func (b *Bot) handleThemeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	theme := options[0].StringValue()
	userID := i.Member.User.ID

	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		b.respondMessage(s, i, "エラーが発生しました")
		return
	}

	err = b.db.UpdateUserSettings(userID, user.Difficulty, theme)
	if err != nil {
		b.respondMessage(s, i, "設定の更新に失敗しました")
		return
	}

	b.respondMessage(s, i, fmt.Sprintf("✅ テーマを「%s」に設定しました！", theme))
}

// handleStatsCommand shows user statistics
func (b *Bot) handleStatsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	stats, err := b.db.GetUserStats(userID)
	if err != nil {
		b.respondMessage(s, i, "統計の取得に失敗しました")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "📊 あなたの学習統計",
		Color: 0x00D4AA,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "総回答数", Value: fmt.Sprintf("%d 問", stats.TotalAnswers), Inline: true},
			{Name: "平均スコア", Value: fmt.Sprintf("%.1f 点", stats.AverageScore), Inline: true},
			{Name: "最高スコア", Value: fmt.Sprintf("%d 点", stats.HighestScore), Inline: true},
			{Name: "今日の回答", Value: fmt.Sprintf("%d 問", stats.AnswersToday), Inline: true},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleSettingsCommand shows current settings
func (b *Bot) handleSettingsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		b.respondMessage(s, i, "設定の取得に失敗しました")
		return
	}

	difficultyLabel := map[string]string{
		"beginner":     "初級",
		"intermediate": "中級",
		"advanced":     "上級",
	}[user.Difficulty]

	embed := &discordgo.MessageEmbed{
		Title: "⚙️ 現在の設定",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "難易度", Value: difficultyLabel, Inline: true},
			{Name: "テーマ", Value: user.Theme, Inline: true},
			{Name: "定期出題", Value: map[bool]string{true: "ON", false: "OFF"}[user.ScheduleEnabled], Inline: true},
		},
	}

	components := b.createSettingsButtons()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// Helper functions

func (b *Bot) respondMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func (b *Bot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

func (b *Bot) createQuizEmbed(questionID int64, japanese, theme, difficulty string) *discordgo.MessageEmbed {
	difficultyLabel := map[string]string{
		"beginner":     "初級",
		"intermediate": "中級",
		"advanced":     "上級",
	}[difficulty]

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📝 英作文問題 #%d", questionID),
		Description: fmt.Sprintf("「%s」", japanese),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "🎯 テーマ", Value: theme, Inline: true},
			{Name: "📊 難易度", Value: difficultyLabel, Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "💡 このメッセージに返信して回答してください！",
		},
	}
}

func (b *Bot) createQuizButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "📝 次の問題",
					Style:    discordgo.PrimaryButton,
					CustomID: "quiz_next",
				},
				discordgo.Button{
					Label:    "⚙️ 設定",
					Style:    discordgo.SecondaryButton,
					CustomID: "settings_open",
				},
				discordgo.Button{
					Label:    "📊 統計",
					Style:    discordgo.SecondaryButton,
					CustomID: "stats_show",
				},
			},
		},
	}
}

func (b *Bot) createSettingsButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "difficulty_select",
					Placeholder: "難易度を選択",
					Options: []discordgo.SelectMenuOption{
						{Label: "初級", Value: "beginner", Description: "シンプルな文法、基本語彙"},
						{Label: "中級", Value: "intermediate", Description: "複文、一般的な表現"},
						{Label: "上級", Value: "advanced", Description: "複雑な文法、慣用句"},
					},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🎯 テーマ変更",
					Style:    discordgo.PrimaryButton,
					CustomID: "theme_modal",
				},
				discordgo.Button{
					Label:    "⏰ スケジュール切替",
					Style:    discordgo.SecondaryButton,
					CustomID: "schedule_toggle",
				},
			},
		},
	}
}
