package game

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Function to render the game state
func (g *Game) RenderEmbed() *discordgo.InteractionResponseData {
	switch g.State {
	case Lobby: // Lobby / start of game
		components := []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Start",
						Style:    discordgo.SuccessButton,
						CustomID: StartButton,
						Disabled: len(g.Players) < 2 || g.State != Lobby,
					},
					&discordgo.Button{
						Label:    "Join",
						Style:    discordgo.PrimaryButton,
						CustomID: JoinButton,
					},
					&discordgo.Button{
						Label:    "Leave",
						Style:    discordgo.SecondaryButton,
						CustomID: LeaveButton,
					},
					&discordgo.Button{
						Label:    "End Game",
						Style:    discordgo.DangerButton,
						CustomID: EndButton,
					},
				},
			},
		}

		// Create the player list as a string (user names or user IDs)
		var playerNames []string
		for _, player := range g.Players {
			// Replace player.UserID with player.Name if you want to display usernames instead of user IDs
			playerNames = append(playerNames, "<@"+player.User.ID+">") // Mention the user using the Discord format
		}

		// Join player names into a string with each player on a new line
		playerList := strings.Join(playerNames, "\n")

		embed := &discordgo.MessageEmbed{
			Title:       "UNO Game Lobby",
			Description: "Welcome to the UNO game lobby! Press 'Join' to join the game, or the host can press 'Start' to begin.",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Players",
					Value:  playerList,
					Inline: false,
				},
			},
			Color: 0x00ff00,
			Image: &discordgo.MessageEmbedImage{
				URL: Cards[0].Link,
			},
		}

		return &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		}
	case Playing: // Playing
		components := []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "UNO!",
						Style:    discordgo.PrimaryButton,
						CustomID: UNOButton,
						Disabled: !g.UNO,
					},
					&discordgo.Button{
						Label:    "View Cards",
						Style:    discordgo.SuccessButton,
						CustomID: ViewCardsButton,
					},
					&discordgo.Button{
						Label:    "Leave",
						Style:    discordgo.SecondaryButton,
						CustomID: LeaveButton,
					},
					&discordgo.Button{
						Label:    "End Game",
						Style:    discordgo.DangerButton,
						CustomID: EndButton,
					},
				},
			},
		}

		// Create the player list as a string (user names or user IDs)
		var playerNames []string
		for _, player := range g.Players {
			if player.User.ID == g.GetCurrentPlayer().User.ID {
				// Replace player.UserID with player.Name if you want to display usernames instead of user IDs
				playerNames = append(playerNames, fmt.Sprintf(">**"+player.User.Username+"**: %d", len(player.Hand))) // Mention the user using the Discord format
				continue
			}
			// Replace player.UserID with player.Name if you want to display usernames instead of user IDs
			playerNames = append(playerNames, fmt.Sprintf(""+player.User.Username+": %d", len(player.Hand))) // Mention the user using the Discord format
		}

		// Join player names into a string with each player on a new line
		playerList := strings.Join(playerNames, "\n")

		embed := &discordgo.MessageEmbed{
			Title:       "It's " + g.GetCurrentPlayer().User.Username + " turn!",
			Description: "Current card is: ",
			Color:       0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Players",
					Value:  playerList,
					Inline: false,
				},
			},
			Image: &discordgo.MessageEmbedImage{
				URL: g.TopCard().Link,
			},
		}

		return &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		}
	default:
		return &discordgo.InteractionResponseData{
			Content: "Default",
		}
	}
}

// Helper function to create an ephemeral response
func ephemeralResponse(content string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
		Type: discordgo.InteractionResponseChannelMessageWithSource,
	}
}

// Helper function to create a simple embed
func simpleEmbed(title, description string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: description,
					Color:       0xFF0000,
				},
			},
		},
	}
}

func renderPlayerHand(g *Game, playerID string) *discordgo.InteractionResponseData {
	player := g.GetPlayer(playerID)

	desc := "It's your turn!"
	if g.GetCurrentPlayer().User.ID != playerID {
		desc = fmt.Sprintf("It's %s turn", g.GetCurrentPlayer().User.Username)
	}

	// Create buttons for each card
	var cardButtons []discordgo.MessageComponent
	for _, card := range player.Hand {
		var colorEmoji string
		switch strings.Split(card.Name, "-")[0] {
		case "red":
			colorEmoji = "🟥"
		case "blue":
			colorEmoji = "🟦"
		case "green":
			colorEmoji = "🟩"
		case "yellow":
			colorEmoji = "🟨"
		default:
			colorEmoji = "⬜" // Wild or undefined color
		}

		cardButtons = append(cardButtons, &discordgo.Button{
			Label:    colorEmoji + strings.ToUpper(card.Name),
			Style:    discordgo.PrimaryButton,
			CustomID: "card-" + card.ID,
			Disabled: g.GetCurrentPlayer().User.ID != player.User.ID, // Disable if not player's turn
		})
	}
	// Add the "Draw Card" button separately
	cardButtons = append(cardButtons, &discordgo.Button{
		Label:    "Draw card",
		Style:    discordgo.SecondaryButton,
		CustomID: "draw-card",
		Disabled: g.GetCurrentPlayer().User.ID != player.User.ID,
	})

	// Group buttons into rows (max 5 buttons per row)
	var rows []discordgo.MessageComponent
	for i := 0; i < len(cardButtons); i += 5 {
		end := i + 5
		if end > len(cardButtons) {
			end = len(cardButtons)
		}
		rows = append(rows, &discordgo.ActionsRow{
			Components: cardButtons[i:end],
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("You have **__%d__** cards in your hand", len(player.Hand)),
		Description: desc,
		Color:       0xFF0000, // Red color code
	}

	return &discordgo.InteractionResponseData{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Flags:      discordgo.MessageFlagsEphemeral,
		Components: rows,
	}
}

func (g *Game) RenderUpdate(s *discordgo.Session) {
	// Update the game view
	if g.Interaction != nil {
		_, err := s.InteractionResponseEdit(g.Interaction, &discordgo.WebhookEdit{
			Embeds:     &g.RenderEmbed().Embeds,
			Components: &g.RenderEmbed().Components,
		})
		if err != nil {
			log.Printf("Failed to update game view: %v", err)
		}
	}

	// Update each players cards
	for _, player := range g.Players {
		if player.Interaction == nil {
			continue
		}

		_, err := s.InteractionResponseEdit(player.Interaction, &discordgo.WebhookEdit{
			Embeds:     &renderPlayerHand(g, player.User.ID).Embeds,
			Components: &renderPlayerHand(g, player.User.ID).Components,
		})
		if err != nil {
			log.Printf("Failed to update player hand: %v", err)
		}
	}
}
