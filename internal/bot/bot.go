package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/melophe/Discord-ENG/internal/claude"
	"github.com/melophe/Discord-ENG/internal/config"
	"github.com/melophe/Discord-ENG/internal/db"
)

// Bot represents the Discord bot
type Bot struct {
	session   *discordgo.Session
	config    *config.Config
	db        *db.DB
	claude    *claude.Client
	channelID string
}

// New creates a new Bot instance
func New(cfg *config.Config, database *db.DB, claudeClient *claude.Client) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session:   session,
		config:    cfg,
		db:        database,
		claude:    claudeClient,
		channelID: cfg.Discord.ChannelID,
	}

	// Register handlers
	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onInteractionCreate)
	session.AddHandler(bot.onMessageCreate)

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsDirectMessages

	return bot, nil
}

// Start opens the Discord connection and registers commands
func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}

	// Register slash commands
	if err := b.registerCommands(); err != nil {
		return err
	}

	log.Println("Bot is running!")
	return nil
}

// Stop closes the Discord connection
func (b *Bot) Stop() error {
	return b.session.Close()
}

// Session returns the Discord session
func (b *Bot) Session() *discordgo.Session {
	return b.session
}

// ChannelID returns the configured channel ID
func (b *Bot) ChannelID() string {
	return b.channelID
}

// onReady is called when the bot is ready
func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

// registerCommands registers slash commands
func (b *Bot) registerCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "quiz",
			Description: "Get a new English translation quiz",
		},
		{
			Name:        "theme",
			Description: "Set the quiz theme",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "theme",
					Description: "The theme for questions (e.g., programming, cooking, business)",
					Required:    true,
				},
			},
		},
		{
			Name:        "stats",
			Description: "View your learning statistics",
		},
		{
			Name:        "settings",
			Description: "View or change your settings",
		},
	}

	for _, cmd := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Cannot create command %v: %v", cmd.Name, err)
			return err
		}
	}

	log.Println("Slash commands registered")
	return nil
}
