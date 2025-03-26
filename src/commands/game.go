package commands

import (
	"log"

	"github.com/Ranzz02/uno-discord-bot/src/game"
	"github.com/bwmarrin/discordgo"
)

func CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	commandData := i.ApplicationCommandData()

	switch commandData.Name {
	case StartCMD:
		game := game.NewGame(i)
		if game == nil {
			s.ChannelMessageSend(i.ChannelID, "Error occurred while creating the lobby")
			return
		}

		// Send the lobby message
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: game.RenderEmbed(),
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
		if err != nil {
			log.Printf("Failed to send embed: %v", err)
			s.ChannelMessageSend(i.ChannelID, "Error occurred while creating the lobby: "+err.Error())
			return
		}
	}
}

func ButtonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()

	g := game.FindGame(i.GuildID)
	if g == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "Game ended or crashed, start a new one.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
		return
	}

	// Switch on customID of button
	switch data.CustomID {
	case game.StartButton:
		g.StartGame(s, i)
	case game.JoinButton:
		g.AddPlayer(s, i)
	case game.LeaveButton:
		g.LeaveGame(s, i)
	case game.EndButton:
		g.Delete(s, i)
	case game.UNOButton:
	case game.ViewCardsButton:
		g.ViewCards(s, i)
	default:
		// If the CustomID doesn't match any known button action
		log.Printf("Unknown button action: %s", data.CustomID)
	}

}
