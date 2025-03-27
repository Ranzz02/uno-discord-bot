package game

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Function to render the game state
func (g *Game) RenderEmbed(s *discordgo.Session) *discordgo.InteractionResponseData {
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

		embed := &discordgo.MessageEmbed{
			Title:       "UNO Game Lobby",
			Description: "Welcome to the UNO game lobby! Press 'Join' to join the game, or the host can press 'Start' to begin.",
			Fields:      playersList(g),
			Color:       0x00ff00,
			Image: &discordgo.MessageEmbedImage{
				URL: Cards[0].Link,
			},
		}

		return &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		}
	case Playing: // Playing
		// Check if the top card is a Wild Card
		topCard := g.TopCard()
		var colorDisplay string

		var wildCardColor *discordgo.MessageEmbedField
		if topCard.Type == WildCard || topCard.Type == WildDrawFourCard {
			colorEmojiMap := map[string]string{
				"red":    "🟥",
				"green":  "🟩",
				"blue":   "🟦",
				"yellow": "🟨",
			}

			selectedColor := *g.CurrentColor
			colorEmoji, exists := colorEmojiMap[selectedColor]
			if !exists {
				colorEmoji = "⬜" // Default if color isn't set
			}

			colorDisplay = fmt.Sprintf("%s %s", colorEmoji, strings.ToUpper(selectedColor))

			wildCardColor = &discordgo.MessageEmbedField{
				Name:   "Wild Color:",
				Value:  colorDisplay,
				Inline: true,
			}
		}

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

		// Add wildCardColor
		fields := playersList(g)
		if wildCardColor != nil {
			fields = append(fields, wildCardColor)
		}

		embed := &discordgo.MessageEmbed{
			Title:       "It's " + g.GetCurrentPlayer().User.Username + " turn!",
			Description: fmt.Sprintf("Current card is: **%s**", topCard.Name),
			Color:       0x00ff00,
			Fields:      fields,
			Image: &discordgo.MessageEmbedImage{
				URL: g.TopCard().Link,
			},
		}

		return &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		}
	case EndScreen:
		var winner *Player
		var players []*Player
		for _, player := range g.Players {
			if len(player.Hand) <= 0 {
				winner = player
				continue
			}
			players = append(players, player)
		}

		var playerList string
		for _, player := range players {
			playerList = playerList + fmt.Sprintf("<@%s> **__%d__** cards left!\n", player.User.ID, len(player.Hand))
		}

		embeds :=
			[]*discordgo.MessageEmbed{
				{
					Title:       "UNO Game Ended",
					Description: "Game has come to an end",
					Color:       0x00ff00,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Winner 👑",
							Value:  fmt.Sprintf("<@%s>", winner.User.ID),
							Inline: false,
						},
						{
							Name:   "Players",
							Value:  playerList,
							Inline: false,
						},
					},
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: winner.User.AvatarURL("1024"),
					},
				},
			}

		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Play Again",
						Style:    discordgo.PrimaryButton,
						CustomID: ReplayButton,
					},
				},
			},
		}

		return &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Components: components,
		}
	default:
		return &discordgo.InteractionResponseData{
			Content: "Default",
		}
	}
}

// Helper function to return PlayerList
func playersList(g *Game) []*discordgo.MessageEmbedField {
	// Create the player list as a string (user names or user IDs)
	var playerNames []string
	currentPlayerID := g.GetCurrentPlayer().User.ID

	for _, player := range g.Players {
		playerFormat := fmt.Sprintf("<@%s>: **__%d__**", player.User.ID, len(player.Hand))

		if g.State == Playing && player.User.ID == currentPlayerID {
			// Replace player.UserID with player.Name if you want to display usernames instead of user IDs
			playerFormat = fmt.Sprintf("> 🎯 <@%s>: **__%d__**", player.User.ID, len(player.Hand))
		}
		// Replace player.UserID with player.Name if you want to display usernames instead of user IDs
		playerNames = append(playerNames, playerFormat) // Mention the user using the Discord format
	}

	// Join player names into a string with each player on a new line
	playerList := strings.Join(playerNames, "\n")

	return []*discordgo.MessageEmbedField{
		{
			Name:   "Players",
			Value:  playerList,
			Inline: false,
		},
	}
}

func renderPlayerHand(g *Game, playerID string) *discordgo.InteractionResponseData {
	player := g.GetPlayer(playerID)

	turnTitle := "It's your turn!"
	if g.GetCurrentPlayer().User.ID != playerID {
		turnTitle = fmt.Sprintf("It's %s turn", g.GetCurrentPlayer().User.Username)
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
			Disabled: g.GetCurrentPlayer().User.ID != player.User.ID || !g.CanPlayCard(&card), // Disable if not player's turn
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
		Title:       turnTitle,
		Description: fmt.Sprintf("You have **__%d__** cards in your hand", len(player.Hand)),
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
			Embeds:     &g.RenderEmbed(s).Embeds,
			Components: &g.RenderEmbed(s).Components,
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
