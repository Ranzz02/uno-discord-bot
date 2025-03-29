package game

import "github.com/bwmarrin/discordgo"

type PlayerRole int

const (
	Host PlayerRole = iota
	Normal
)

type Player struct {
	User        *discordgo.User
	Hand        []Card
	Role        PlayerRole
	Interaction *discordgo.Interaction
	Page        int
}

func (p *Player) HasValidPreviousPlay(g *Game) bool {
	for _, card := range p.Hand {
		if g.CanPlayPreviousCard(&card) {
			return true
		}
	}

	return false
}

func (g *Game) NewPlayer(user *discordgo.User, role PlayerRole, initCards int) {
	g.Players = append(g.Players, &Player{
		User: user,
		Hand: DrawCards(g, initCards),
		Role: role,
	})
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
func (g *Game) GetCurrentPlayer() *Player {
	return g.Players[g.CurrentTurn]
}

// Get previous player
func (g *Game) GetPreviousPlayer() *Player {
	if g.Reversed {
		// If the direction is reversed, move backwards.
		return g.Players[(g.CurrentTurn+1+len(g.Players))%len(g.Players)]
	}
	// Otherwise, move forwards.
	return g.Players[(g.CurrentTurn-1)%len(g.Players)]
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
