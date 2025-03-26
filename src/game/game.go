package game

import (
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	StartButton string = "start_button"
	JoinButton  string = "join_button"
	LeaveButton string = "leave_button"
	EndButton   string = "end_button"
	// Playing
	UNOButton       string = "uno_button"
	ViewCardsButton string = "view_cards_button"
	DrawCardAction  string = "draw-card"
)

var (
	games    = map[string]*Game{}
	gamesMux = sync.Mutex{}
)

type GameState int

const (
	Lobby GameState = iota
	Playing
	EndScreen
)

type Game struct {
	ID          string
	Deck        []Card
	DiscardPile []Card
	Players     []*Player
	CurrentTurn int
	Reversed    bool
	UNO         bool
	State       GameState
	Host        string
	Interaction *discordgo.Interaction
}

// Start a new game
func NewGame(i *discordgo.InteractionCreate) *Game {
	id, err := gonanoid.New()
	if err != nil {
		return nil
	}

	game := &Game{
		ID:          id,
		Deck:        GenerateDeck(),
		DiscardPile: []Card{},
		CurrentTurn: 0,
		Reversed:    false,
		State:       Lobby,
		UNO:         false,
		Host:        i.Member.User.ID,
		Interaction: i.Interaction,
	}

	// Shuffle deck
	ShuffleDeck(game.Deck)
	game.TopCard()

	// Select the first card and validate it's not a Wild or Wild Draw Four
	firstCard := game.Deck[0] // Get the first card from the shuffled deck

	// Ensure the first card isn't a Wild or Wild Draw Four card
	for firstCard.Type == WildCard || firstCard.Type == WildDrawFourCard {
		// If it's a Wild or Wild Draw Four, reshuffle the deck and pick a new first card
		ShuffleDeck(game.Deck)
		firstCard = game.Deck[0]
	}

	// Place the first valid card on the discard pile
	game.DiscardPile = append(game.DiscardPile, firstCard)
	game.Deck = game.Deck[1:] // Remove the first card from the deck

	game.Players = append(game.Players, &Player{
		User: i.Member.User,
		Hand: DrawCards(game, 7),
		Role: Host,
	})

	gamesMux.Lock()
	games[i.ChannelID] = game
	gamesMux.Unlock()

	return game
}

// Find a game
func FindGame(gameID string) *Game {
	gamesMux.Lock()
	defer gamesMux.Unlock()

	g, ok := games[gameID]
	if !ok {
		return nil
	}
	return g
}

func (g *Game) TopCard() Card {
	// Check if the discard pile is empty
	if len(g.DiscardPile) == 0 {
		// If the discard pile is empty, draw the first card from the deck
		firstCard := g.Deck[0]

		// Ensure the first card isn't a Wild or Wild Draw Four card
		for firstCard.Type == WildCard || firstCard.Type == WildDrawFourCard {
			// If it's a Wild or Wild Draw Four, reshuffle the deck and pick a new first card
			ShuffleDeck(g.Deck)
			firstCard = g.Deck[0]
		}

		// Place the valid card on the discard pile and remove it from the deck
		g.DiscardPile = append(g.DiscardPile, firstCard)
		g.Deck = g.Deck[1:] // Remove the first card from the deck

		return firstCard
	}

	return g.DiscardPile[len(g.DiscardPile)-1]
}

// Function to check if card can be played
func (g *Game) CanPlayCard(card *Card) bool {
	topCard := g.TopCard()

	if card.Type == WildCard || card.Type == WildDrawFourCard {
		return true
	}

	// Split the card name to get color and type (number)
	cardDetails := strings.Split(card.Name, "-")
	cardColor := cardDetails[0]
	cardType := cardDetails[1]

	// Split the top card to get the color and type
	topCardDetails := strings.Split(topCard.Name, "-")
	topCardColor := topCardDetails[0]
	topCardType := topCardDetails[1]

	// Check if the card can be played (must match either color or type)
	if cardColor == topCardColor || cardType == topCardType {
		return true
	}

	return false
}
