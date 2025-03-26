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

// Get next player
func (g *Game) GetNextPlayer() *Player {
	if g.Reversed {
		// If the direction is reversed, move backwards.
		return g.Players[(g.CurrentTurn-1+len(g.Players))%len(g.Players)]
	}
	// Otherwise, move forwards.
	return g.Players[(g.CurrentTurn+1)%len(g.Players)]
}
