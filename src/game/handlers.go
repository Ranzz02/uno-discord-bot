package game

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Add player to the game
func (g *Game) AddPlayer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if g.State != Lobby {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "This game started already.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
		return
	}

	exists := false
	for _, player := range g.Players {
		if player.User.ID == i.Member.User.ID {
			exists = true
			break
		}
	}

	if exists {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "You are already in the game.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
		return
	}

	g.NewPlayer(i.Member.User, Normal, 7)

	// Respond with the updated embed, rendering the correct buttons
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Data: g.RenderEmbed(s),
		Type: discordgo.InteractionResponseUpdateMessage,
	})

}

// Player leaves the game
func (g *Game) LeaveGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if g.State != Lobby {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "This game started already. Wait till it ends",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: g.RenderEmbed(s),
	})
}

// Start game
func (g *Game) StartGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if g.Host == i.Member.User.ID {
		g.State = Playing
		// Send an update with the embed (you can modify the existing message or send a new one)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: g.RenderEmbed(s),
			Type: discordgo.InteractionResponseUpdateMessage,
		})
	} else if len(g.Players) >= 2 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "Not enough players",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "You are not the host of the game.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
	}
}

// End game early
func (g *Game) Delete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if g.Host == i.Member.User.ID {
		log.Println("Ending game")

		interaction := i.Interaction
		embed := &discordgo.MessageEmbed{
			Title:       "Game ended",
			Description: "Game was deleted by host",
			Color:       0xFF0000, // Red color code
		}

		s.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
			Type: discordgo.InteractionResponseUpdateMessage,
		})

		go func() {
			time.Sleep(2 * time.Second)
			s.InteractionResponseDelete(interaction)
		}()

		delete(games, g.ID)
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: &discordgo.InteractionResponseData{
				Content: "Only host can end game.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
			Type: discordgo.InteractionResponseChannelMessageWithSource,
		})
	}
}

// View card deck
func (g *Game) ViewCards(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := g.GetPlayer(i.Member.User.ID)
	if player == nil {
		return
	}

	player.Interaction = i.Interaction

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: renderPlayerHand(g, player.User.ID),
	})
	if err != nil {
		log.Printf("Failed to create view_cards: %v", err)
	}
}
