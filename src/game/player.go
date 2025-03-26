package game

import "github.com/bwmarrin/discordgo"

type PlayerRole int

const (
	Host PlayerRole = iota
	Normal
)

type Player struct {
	User *discordgo.User
	Hand []Card
	Role PlayerRole
}
