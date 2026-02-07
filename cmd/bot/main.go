package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/melophe/Discord-ENG/internal/bot"
	"github.com/melophe/Discord-ENG/internal/claude"
	"github.com/melophe/Discord-ENG/internal/config"
	"github.com/melophe/Discord-ENG/internal/db"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize Claude client
	claudeClient := claude.NewClient(cfg.Claude.APIKey, cfg.Claude.Model)

	// Initialize bot
	discordBot, err := bot.New(cfg, database, claudeClient)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Start the bot
	if err := discordBot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
	defer discordBot.Stop()

	// Start scheduler for periodic quizzes
	scheduler := bot.NewScheduler(discordBot, cfg.Schedule.IntervalMinutes)
	scheduler.Start()
	defer scheduler.Stop()

	log.Println("Bot is now running. Press CTRL+C to exit.")

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}
