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
	g.UNO = false
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
		// TODO: Add error response
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
		if len(g.Players) == 2 {
			// 2 Player reverse works as skip
			g.NextTurn()
		}
		g.NextTurn()
	case DrawTwoCard:
		// Force the next player to draw two cards and skip their turn.
		nextPlayer := g.GetNextPlayer()
		drawnCards := DrawCards(g, 2)
		nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)
		g.NextTurn()
		g.NextTurn()
	case WildCard:
		// Block until color selection is completed
		selectedColor := g.ChangeColor(s, i)
		g.CurrentColor = &selectedColor

		// Move to the next player's turn.
		g.NextTurn()
	case WildDrawFourCard:
		// Block until color selection is completed
		selectedColor := g.ChangeColor(s, i)
		g.CurrentColor = &selectedColor

		// Challenge draw four
		challenged := g.ChallengeChoice(s, i)

		log.Printf("Got answer for challenge: %v", challenged)

		if challenged {
			player := g.GetCurrentPlayer()

			if player.HasValidPreviousPlay(g) { // If not only valid card draw 4
				drawnCards := DrawCards(g, 4)
				player.Hand = append(player.Hand, drawnCards...)
			} else { // Punish next player and draw 6 cards
				nextPlayer := g.GetNextPlayer()
				drawnCards := DrawCards(g, 6)
				nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)

				// The challenger loses the challenge and their turn is skipped
				g.NextTurn()
			}
		} else {
			nextPlayer := g.GetNextPlayer()
			drawnCards := DrawCards(g, 4)
			nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)

			// The challenger loses the challenge and their turn is skipped
			g.NextTurn()
		}

		// Move to the next player's turn.
		g.NextTurn()
	}

	if len(player.Hand) == 1 {
		g.UNO = true
	}

	if len(player.Hand) == 0 {
		g.EndGame(s, player)
	}

	// Force a render update after any interaction
	g.RenderUpdate(s)
}

func (g *Game) DrawCard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := g.GetCurrentPlayer()

	// Draw one card
	card := DrawCards(g, 1)
	player.Hand = append(player.Hand, card...)

	// Go to next turn
	g.NextTurn()

	// Force a render update after any interaction
	g.RenderUpdate(s)
}
