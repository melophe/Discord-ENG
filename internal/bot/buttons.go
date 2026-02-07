package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// handleComponentInteraction handles button and select menu interactions
func (b *Bot) handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	switch customID {
	case "quiz_next":
		b.handleNextQuizButton(s, i)
	case "settings_open":
		b.handleSettingsButton(s, i)
	case "stats_show":
		b.handleStatsButton(s, i)
	case "difficulty_select":
		b.handleDifficultySelect(s, i)
	case "theme_modal":
		b.handleThemeModalButton(s, i)
	case "schedule_toggle":
		b.handleScheduleToggle(s, i)
	}
}

// handleNextQuizButton generates a new quiz
func (b *Bot) handleNextQuizButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	userID := i.Member.User.ID
	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}

	ctx := context.Background()
	japanese, err := b.claude.GenerateQuestion(ctx, user.Theme, user.Difficulty)
	if err != nil {
		log.Printf("Error generating question: %v", err)
		msg := "問題の生成に失敗しました"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	questionID, _ := b.db.SaveQuestion(japanese, user.Difficulty, user.Theme)

	embed := b.createQuizEmbed(questionID, japanese, user.Theme, user.Difficulty)
	components := b.createQuizButtons()

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
}

// handleSettingsButton shows settings
func (b *Bot) handleSettingsButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		b.respondComponentMessage(s, i, "設定の取得に失敗しました")
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
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleStatsButton shows statistics
func (b *Bot) handleStatsButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	stats, err := b.db.GetUserStats(userID)
	if err != nil {
		b.respondComponentMessage(s, i, "統計の取得に失敗しました")
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
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleDifficultySelect handles difficulty selection
func (b *Bot) handleDifficultySelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	values := i.MessageComponentData().Values
	if len(values) == 0 {
		return
	}

	difficulty := values[0]
	userID := i.Member.User.ID

	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		b.respondComponentMessage(s, i, "エラーが発生しました")
		return
	}

	err = b.db.UpdateUserSettings(userID, difficulty, user.Theme)
	if err != nil {
		b.respondComponentMessage(s, i, "設定の更新に失敗しました")
		return
	}

	difficultyLabel := map[string]string{
		"beginner":     "初級",
		"intermediate": "中級",
		"advanced":     "上級",
	}[difficulty]

	b.respondComponentMessage(s, i, fmt.Sprintf("✅ 難易度を「%s」に設定しました！", difficultyLabel))
}

// handleThemeModalButton opens the theme input modal
func (b *Bot) handleThemeModalButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "theme_modal_submit",
			Title:    "テーマを設定",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "theme_input",
							Label:       "テーマ",
							Style:       discordgo.TextInputShort,
							Placeholder: "例: プログラミング、料理、旅行",
							Required:    true,
							MinLength:   1,
							MaxLength:   50,
						},
					},
				},
			},
		},
	})
}

// handleScheduleToggle toggles the schedule setting
func (b *Bot) handleScheduleToggle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		b.respondComponentMessage(s, i, "エラーが発生しました")
		return
	}

	newEnabled := !user.ScheduleEnabled
	err = b.db.UpdateUserSchedule(userID, newEnabled)
	if err != nil {
		b.respondComponentMessage(s, i, "設定の更新に失敗しました")
		return
	}

	status := "OFF"
	if newEnabled {
		status = "ON"
	}
	b.respondComponentMessage(s, i, fmt.Sprintf("✅ 定期出題を %s にしました！", status))
}

// handleModalSubmit handles modal form submissions
func (b *Bot) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ModalSubmitData().CustomID != "theme_modal_submit" {
		return
	}

	var theme string
	for _, row := range i.ModalSubmitData().Components {
		for _, comp := range row.(*discordgo.ActionsRow).Components {
			if input, ok := comp.(*discordgo.TextInput); ok && input.CustomID == "theme_input" {
				theme = input.Value
			}
		}
	}

	if theme == "" {
		b.respondComponentMessage(s, i, "テーマを入力してください")
		return
	}

	userID := i.Member.User.ID
	user, err := b.db.GetOrCreateUser(userID)
	if err != nil {
		b.respondComponentMessage(s, i, "エラーが発生しました")
		return
	}

	err = b.db.UpdateUserSettings(userID, user.Difficulty, theme)
	if err != nil {
		b.respondComponentMessage(s, i, "設定の更新に失敗しました")
		return
	}

	b.respondComponentMessage(s, i, fmt.Sprintf("✅ テーマを「%s」に設定しました！", theme))
}

func (b *Bot) respondComponentMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
