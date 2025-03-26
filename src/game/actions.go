package game

import (
	"fmt"

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
		// Allow the current player to change the color. (This is a simple placeholder for the logic.)
		// You might want to implement a way to let the player choose the color.
		// For now, we'll just set a random color for demonstration.
		g.ChangeColor("red") // Example, change to red.

		// Move to the next player's turn.
		g.NextTurn()
	case WildDrawFourCard:
		// Force the next player to draw four cards and skip their turn.
		nextPlayer := g.GetNextPlayer()
		drawnCards := DrawCards(g, 4)
		nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)
		g.ChangeColor("blue") // Example, change to blue.

		// Skip the next player's turn after they draw.
		g.NextTurn()
		g.NextTurn()
	}

	if len(player.Hand) == 1 {
		g.UNO = true
	}
}

// Change the current color for Wild and WildDrawFour cards
func (g *Game) ChangeColor(newColor string) {
	// Here you can implement logic to set the color for the game
	// For example, you can store it in the game state and check the current color when players play cards.
	// This is just a placeholder.
	fmt.Println("Changing color to", newColor)
}

func (g *Game) DrawCard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := g.GetCurrentPlayer()

	// Draw one card
	card := DrawCards(g, 1)
	player.Hand = append(player.Hand, card...)

	// Go to next turn
	g.NextTurn()
}
