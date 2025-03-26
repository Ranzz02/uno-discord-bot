package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	StartCMD string = "uno"
)

var (
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        StartCMD,
			Description: "Start a new uno game",
		},
	}
)

func RegisterCommands(s *discordgo.Session, guildID string) {
	created, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, guildID, Commands)
	if err != nil {
		log.Fatalf("Error registering commands: %v", err)
		return

	}
	log.Printf("✅  %d/%d commands created or overwritten!", len(created), len(Commands))
}
