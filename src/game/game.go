package game

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

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
}

// Shuffle deck of cards
func ShuffleDeck(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

// Draw cards from deck
func DrawCards(deck *[]Card, num int) []Card {
	if len(*deck) < num {
		num = len(*deck)
	}
	drawn := (*deck)[:num]
	*deck = (*deck)[num:]
	return drawn
}

// Find a game
func FindGame(guildId string) *Game {
	gamesMux.Lock()
	defer gamesMux.Unlock()

	g, ok := games[guildId]
	if !ok {
		return nil
	}
	return g
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
	}

	ShuffleDeck(game.Deck)

	game.Players = append(game.Players, &Player{
		User: i.Member.User,
		Hand: DrawCards(&game.Deck, 7),
		Role: Host,
	})

	gamesMux.Lock()
	games[i.GuildID] = game
	gamesMux.Unlock()

	return game
}

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
			if player.User.ID == g.CurrentPlayer().User.ID {
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
			Title:       "It's " + g.CurrentPlayer().User.Username + " turn!",
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
				URL: g.Deck[0].Link,
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

// Get player with id
func (g *Game) GetPlayer(userId string) *Player {
	for _, player := range g.Players {
		if player.User.ID == userId {
			return player
		}
	}
	return nil
}

// Get current player
func (g *Game) CurrentPlayer() *Player {
	return g.Players[g.CurrentTurn]
}

// Get next player
func (g *Game) GetNextPlayer() *Player {
	if g.Reversed {
		// If the direction is reversed, move backwards.
		return g.Players[(g.CurrentTurn-1+len(g.Players))%len(g.Players)]
	}
	// Otherwise, move forwards.
	return g.Players[(g.CurrentTurn+1)%len(g.Players)]
}

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
func (g *Game) PlayCard(card *Card) error {
	player := g.CurrentPlayer()

	newHand := []Card{}
	for _, handCard := range player.Hand {
		if handCard.Name == card.Name {
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
		drawnCards := DrawCards(&g.Deck, 2)
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
		drawnCards := DrawCards(&g.Deck, 4)
		nextPlayer.Hand = append(nextPlayer.Hand, drawnCards...)
		g.ChangeColor("blue") // Example, change to blue.

		// Skip the next player's turn after they draw.
		g.NextTurn()
		g.NextTurn()
	}

	return nil
}

// Change the current color for Wild and WildDrawFour cards
func (g *Game) ChangeColor(newColor string) {
	// Here you can implement logic to set the color for the game
	// For example, you can store it in the game state and check the current color when players play cards.
	// This is just a placeholder.
	fmt.Println("Changing color to", newColor)
}

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

	g.Players = append(g.Players, &Player{
		User: i.Member.User,
		Hand: DrawCards(&g.Deck, 7),
		Role: Normal,
	})

	// Respond with the updated embed, rendering the correct buttons
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Data: g.RenderEmbed(),
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
		Data: g.RenderEmbed(),
	})
}

// Start game
func (g *Game) StartGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if g.Host == i.Member.User.ID {
		g.State = Playing
		// Send an update with the embed (you can modify the existing message or send a new one)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Data: g.RenderEmbed(),
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

	desc := "It's your turn!"
	if &player.User.ID != &i.Member.User.ID {
		desc = fmt.Sprintf("It's %s turn", g.CurrentPlayer().User.Username)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("You have %d", len(player.Hand)),
		Description: desc,
		Color:       0xFF0000, // Red color code
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
