package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	StartCMD string = "uno"
	HelpCMD  string = "help"
)

var (
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        StartCMD,
			Description: "Start a new uno game",
		},
		{
			Name:        HelpCMD,
			Description: "Help with how to use the uno bot",
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
