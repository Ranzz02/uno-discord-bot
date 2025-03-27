package game

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// Moves turn to the next player
func (g *Game) NextTurn() {
	if g.Reversed {
		g.CurrentTurn--
		if g.CurrentTurn < 0 {
			g.CurrentTurn = len(g.Players) - 1
		}
	} else {
		g.CurrentTurn++
		if g.CurrentTurn >= len(g.Players) {
			g.CurrentTurn = 0
		}
	}
}

// Tries to play a players card on their turn
func (g *Game) PlayCard(s *discordgo.Session, i *discordgo.InteractionCreate, cardID string) {
	player := g.GetCurrentPlayer()

	var card *Card
	for _, c := range player.Hand {
		if c.ID == cardID {
			card = &c
			break
		}
	}

	if card == nil {
		return
	}

	if !g.CanPlayCard(card) {
		return
	}

	newHand := []Card{}
	for _, handCard := range player.Hand {
		if handCard.ID == card.ID {
			g.DiscardPile = append(g.DiscardPile, handCard)
			continue
		}
		newHand = append(newHand, handCard)
	}
	player.Hand = newHand

	switch card.Type {
	case NumberCard:
		g.NextTurn()
	case SkipCard:
		g.NextTurn()
		g.NextTurn()
	case ReverseCard:
		g.Reversed = !g.Reversed
		g.NextTurn()
	case DrawTwoCard:
		// Force the next player to draw two cards and skip their turn.
		nextPlayer := g.GetNextPlayer()
		drawnCards := DrawCards(g, 2)
		nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)
		g.NextTurn()
		g.NextTurn()
	case WildCard:
		g.ChangeColor(s, i)
		// Move to the next player's turn.
		g.NextTurn()
	case WildDrawFourCard:
		// Force the next player to draw four cards and skip their turn.
		g.ChangeColor(s, i)
		nextPlayer := g.GetNextPlayer()
		drawnCards := DrawCards(g, 4)
		nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)
		// Skip the next player's turn after they draw.
		g.NextTurn()
		g.NextTurn()
	}

	if len(player.Hand) == 1 {
		g.UNO = true
	}
}

// Change the current color for Wild and WildDrawFour cards
func (g *Game) ChangeColor(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Send an ephemeral message asking the player to select a color
	colorPrompt := "Please select a color for the Wild card!"

	// Send the message to the player, this will be ephemeral (only visible to the player)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Select color",
					Description: colorPrompt,
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "🟥 Red",
							Style:    discordgo.SecondaryButton,
							CustomID: "color_red",
						},
						&discordgo.Button{
							Label:    "🟩 Green",
							Style:    discordgo.SecondaryButton,
							CustomID: "color_green",
						},
						&discordgo.Button{
							Label:    "🟦 Blue",
							Style:    discordgo.SecondaryButton,
							CustomID: "color_blue",
						},
						&discordgo.Button{
							Label:    "🟨 Yellow",
							Style:    discordgo.SecondaryButton,
							CustomID: "color_yellow",
						},
					},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
		Type: discordgo.InteractionResponseChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error sending color selection message: %v", err)
		return
	}

	// Wait for the user to react with one of the color emojis
	g.WaitForColorSelection(s, i)
}

func (g *Game) WaitForColorSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.AddHandlerOnce(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}

		data := i.MessageComponentData()

		var selectedColor string
		switch data.CustomID {
		case "🟥":
			selectedColor = "color_red"
		case "🟩":
			selectedColor = "color_green"
		case "🟦":
			selectedColor = "color_blue"
		case "🟨":
			selectedColor = "color_yellow"
		default:
			// If the reaction is not valid, ignore it
			return
		}

		// Update the game state with the selected color
		g.CurrentColor = &selectedColor

		g.NextTurn()
	})
}

func (g *Game) DrawCard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := g.GetCurrentPlayer()

	// Draw one card
	card := DrawCards(g, 1)
	player.Hand = append(player.Hand, card...)

	// Go to next turn
	g.NextTurn()
}
