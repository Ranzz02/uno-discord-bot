package commands

import (
	"log"
	"strings"

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
			Data: game.RenderEmbed(s),
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

	g := game.FindGame(i.ChannelID)
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
	switch {
	case data.CustomID == game.StartButton:
		// Start game
		g.StartGame(s, i)
	case data.CustomID == game.JoinButton:
		// Add player to game
		g.AddPlayer(s, i)
	case data.CustomID == game.LeaveButton:
		// Leave game
		g.LeaveGame(s, i)
	case data.CustomID == game.EndButton:
		// Pre-End game
		g.Delete(s, i)
	case data.CustomID == game.UNOButton:
		// Call UNO
	case data.CustomID == game.ReplayButton:
		// Replay button
	case data.CustomID == game.ViewCardsButton:
		// Create a view of players hand
		g.ViewCards(s, i)
	case data.CustomID == game.DrawCardAction: // Draw one card from deck
		// Draw a card from the pile
		g.DrawCard(s, i)
	case strings.HasPrefix(data.CustomID, "card-"):
		// Play card
		cardID := strings.TrimPrefix(data.CustomID, "card-")
		g.PlayCard(s, i, cardID)
	case data.CustomID == game.PreviousButton: // Previous hand
		player := g.GetPlayer(i.Member.User.ID)
		if player != nil && player.Page > 0 {
			player.Page--
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: g.RenderPlayerHand(player.User.ID),
			})
		}
	case data.CustomID == game.NextButton: // Next hand
		player := g.GetPlayer(i.Member.User.ID)
		if player != nil && player.Page < game.MAX_CARDS_PER_PAGE-1 {
			player.Page++
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: g.RenderPlayerHand(player.User.ID),
			})
		}
	default:
		// If the CustomID doesn't match any known button action
		log.Printf("Unknown button action: %s", data.CustomID)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Data: nil,
		Type: discordgo.InteractionResponseUpdateMessage,
	})
}
