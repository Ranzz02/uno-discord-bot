package game

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	StartButton  string = "start_button"
	JoinButton   string = "join_button"
	LeaveButton  string = "leave_button"
	EndButton    string = "end_button"
	ReplayButton string = "replay_button"
	// Playing
	UNOButton           string = "uno_button"
	ViewCardsButton     string = "view_cards_button"
	DrawCardAction      string = "draw-card"
	KeepCardAction      string = "keep-card"
	PlayDrawnCardAction string = "play-drawn-card"
	// Challenge buttons
	ChallengeButton       string = "challenge_button"
	ChallengeIgnoreButton string = "challenge_ignore"
	// Pagination buttons
	PreviousButton string = "previous_button"
	NextButton     string = "next_button"
)

const (
	MAX_CARDS_PER_PAGE int = 15
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
	ID            string
	Deck          []Card
	DiscardPile   []Card
	Players       []*Player
	CurrentTurn   int
	Reversed      bool
	UNO           bool
	State         GameState
	Host          string
	Interaction   *discordgo.Interaction
	ColorData     ColorData
	ChallengeData ChallengeData
	KeepCardData  KeepCardData
	Winner        *Player
}

type ColorData struct {
	ColorResponse chan string
	CurrentColor  *string
	User          string
}

type ChallengeData struct {
	ChallengeResponse chan bool
	User              string
}

type KeepCardData struct {
	KeepResponse chan bool
	User         string
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
		ColorData: ColorData{
			ColorResponse: make(chan string, 5),
		},
		ChallengeData: ChallengeData{
			ChallengeResponse: make(chan bool, 1),
		},
		KeepCardData: KeepCardData{
			KeepResponse: make(chan bool, 1),
		},
	}

	// Shuffle deck
	ShuffleDeck(game.Deck)
	game.TopCard()

	// Select the first card and validate it's not a Wild or Wild Draw Four
	firstCard := game.Deck[0] // Get the first card from the shuffled deck

	// Ensure the first card isn't a Wild or Wild Draw Four card
	for firstCard.Type != NumberCard {
		// If it's a Wild or Wild Draw Four, reshuffle the deck and pick a new first card
		ShuffleDeck(game.Deck)
		firstCard = game.Deck[0]
	}

	// Place the first valid card on the discard pile
	game.DiscardPile = append(game.DiscardPile, firstCard)
	game.Deck = game.Deck[1:] // Remove the first card from the deck

	// Add host to game
	game.NewPlayer(i.Member.User, Host, 7)

	gamesMux.Lock()
	games[i.ChannelID] = game
	gamesMux.Unlock()

	return game
}

// End game with winner
func (g *Game) EndGame(s *discordgo.Session, player *Player) {
	g.State = EndScreen
	g.Winner = player

	// Delete view hands
	for _, player := range g.Players {
		if player.Interaction != nil {
			s.InteractionResponseDelete(player.Interaction)
		}
	}

	// Remove game from games
	gamesMux.Lock()
	defer gamesMux.Unlock()
	delete(games, g.ID)

	// Update UI
	g.RenderUpdate(s)
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

	// If wild card
	if topCardColor == "wild" && g.ColorData.CurrentColor != nil && *g.ColorData.CurrentColor == cardColor {
		return true
	}

	// Check if the card can be played (must match either color or type)
	if cardColor == topCardColor || cardType == topCardType {
		return true
	}

	return false
}

// Function to check if card can be played
func (g *Game) CanPlayPreviousCard(card *Card) bool {
	topCard := g.DiscardPile[len(g.DiscardPile)-2]

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

	// If wild card
	if topCardColor == "wild" && g.ColorData.CurrentColor != nil && *g.ColorData.CurrentColor == cardColor {
		return true
	}

	// Check if the card can be played (must match either color or type)
	if cardColor == topCardColor || cardType == topCardType {
		return true
	}

	return false
}

// Change the current color for Wild and WildDrawFour cards
func (g *Game) ChangeColor(s *discordgo.Session, i *discordgo.InteractionCreate) string {
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
		return ""
	}

	// Wait for the user to react with one of the color emojis
	selectedColor := g.WaitForColorSelection(s, i)
	g.ColorData.CurrentColor = &selectedColor
	return selectedColor
}

func (g *Game) WaitForColorSelection(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	g.ColorData.User = i.Member.User.ID

	select {
	case selectedColor := <-g.ColorData.ColorResponse:
		return selectedColor
	case <-time.After(30 * time.Second):
		return "red"
	}
}

func (g *Game) ChallengeChoice(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	nextPlayer := g.GetNextPlayer()
	g.ChallengeData.User = nextPlayer.User.ID

	// Send the message to the player, this will be ephemeral (only visible to the player)
	_, err := s.InteractionResponseEdit(nextPlayer.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       "Challenge wild draw four!",
				Description: "Do you want to challenge the Wild Draw Four?",
			},
		},
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Challenge",
						Style:    discordgo.DangerButton,
						CustomID: ChallengeButton,
					},
					&discordgo.Button{
						Label:    "Ignore",
						Style:    discordgo.SecondaryButton,
						CustomID: ChallengeIgnoreButton,
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Error sending color selection message: %v", err)
		return false
	}

	// Wait for the user to react with one of the color emojis
	return g.WaitForChallengeSelection(s, i)
}

func (g *Game) WaitForChallengeSelection(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	// Wait for a response or timeout
	select {
	case selectedChoice := <-g.ChallengeData.ChallengeResponse:
		log.Printf("Challenge choice made: %v", selectedChoice)
		return selectedChoice // Player responded
	case <-time.After(30 * time.Second):
		return false // Timeout occurred, default choice (false)
	}
}

func (g *Game) HandleDrawCard(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	player := g.GetCurrentPlayer()

	// Determine the color of the embed based on the card
	var embedColor int
	switch strings.Split(player.LastDrawnCard.Name, "-")[0] {
	case "red":
		embedColor = 0xFF0000
	case "blue":
		embedColor = 0x0000FF
	case "green":
		embedColor = 0x00FF00
	case "yellow":
		embedColor = 0xFFFF00
	default:
		embedColor = 0xFFFFFF // White for wild cards
	}

	// Create an embed showing the drawn card
	embed := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("You drew a **%s**!", player.LastDrawnCard.Name),
		Description: "Do you want to play it or keep it?",
		Color:       embedColor,
		Image: &discordgo.MessageEmbedImage{
			URL: player.LastDrawnCard.Link,
		},
	}

	// Ask the player to decide
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "Play card",
							Style:    discordgo.SuccessButton,
							CustomID: PlayDrawnCardAction,
							Disabled: !g.CanPlayCard(player.LastDrawnCard),
						},
						&discordgo.Button{
							Label:    "Keep",
							Style:    discordgo.SecondaryButton,
							CustomID: KeepCardAction,
						},
					},
				},
			},
		},
		Type: discordgo.InteractionResponseUpdateMessage,
	})
	if err != nil {
		log.Printf("Failed to ask to keep card: %v", err)
		return false
	}

	return g.WaitForKeepCard(s, i)
}

// Wait for response or timeout after 10 seconds
func (g *Game) WaitForKeepCard(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	select {
	case selectedChoice := <-g.KeepCardData.KeepResponse:
		return selectedChoice
	case <-time.After(10 * time.Second):
		return true
	}
}
