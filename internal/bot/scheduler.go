package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Scheduler handles periodic quiz posting
type Scheduler struct {
	bot      *Bot
	interval time.Duration
	stop     chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(bot *Bot, intervalMinutes int) *Scheduler {
	return &Scheduler{
		bot:      bot,
		interval: time.Duration(intervalMinutes) * time.Minute,
		stop:     make(chan struct{}),
	}
}

// Start begins the periodic quiz posting
func (s *Scheduler) Start() {
	go s.run()
	log.Printf("Scheduler started (interval: %v)", s.interval)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stop)
	log.Println("Scheduler stopped")
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.postScheduledQuiz()
		case <-s.stop:
			return
		}
	}
}

// postScheduledQuiz posts a quiz to the configured channel
func (s *Scheduler) postScheduledQuiz() {
	log.Println("Posting scheduled quiz...")

	// Use default settings for scheduled quizzes
	theme := "日常会話"
	difficulty := "intermediate"

	ctx := context.Background()
	japanese, err := s.bot.claude.GenerateQuestion(ctx, theme, difficulty)
	if err != nil {
		log.Printf("Error generating scheduled question: %v", err)
		return
	}

	questionID, err := s.bot.db.SaveQuestion(japanese, difficulty, theme)
	if err != nil {
		log.Printf("Error saving scheduled question: %v", err)
		return
	}

	embed := s.createScheduledQuizEmbed(questionID, japanese, theme, difficulty)
	components := s.bot.createQuizButtons()

	_, err = s.bot.Session().ChannelMessageSendComplex(s.bot.ChannelID(), &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})
	if err != nil {
		log.Printf("Error posting scheduled quiz: %v", err)
		return
	}

	log.Printf("Scheduled quiz #%d posted successfully", questionID)
}

// createScheduledQuizEmbed creates an embed for scheduled quizzes
func (s *Scheduler) createScheduledQuizEmbed(questionID int64, japanese, theme, difficulty string) *discordgo.MessageEmbed {
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
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
