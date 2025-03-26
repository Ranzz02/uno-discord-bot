package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ranzz02/uno-discord-bot/src/commands"
	"github.com/Ranzz02/uno-discord-bot/src/config"
	"github.com/bwmarrin/discordgo"
)

var Bot *discordgo.Session

func InitBot() {
	var err error
	Bot, err = discordgo.New("Bot " + config.Conf.Token)
	if err != nil {
		log.Fatalf("Failed to start bot: %v", err)
		panic(err)
	}

	Bot.AddHandler(commands.CommandHandler)
	Bot.AddHandler(commands.ButtonHandler)

	Bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = Bot.Open()
	if err != nil {
		log.Printf("Error opening connection: %v", err)
		return
	}

	// Set bots status
	err = Bot.UpdateListeningStatus("/uno")
	if err != nil {
		log.Fatalf("Error updating status %v", err)
	}

	commands.RegisterCommands(Bot, "")

	log.Println("Bot is now running. Press CTRL+C to exit.")
	gracefulShutdown()
}

// gracefulShutdown listens for interrupt signals to cleanly shut down the bot.
func gracefulShutdown() {
	// Create a channel to listen for interrupt signals (Ctrl+C or SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for the signal to stop
	<-stop

	// Close the Discord session
	log.Println("Shutting down bot...")
	if err := Bot.Close(); err != nil {
		log.Printf("Error closing the connection: %v", err)
	}
	log.Println("Bot shut down successfully.")
}
