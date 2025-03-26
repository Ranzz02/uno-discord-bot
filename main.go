package main

import (
	"github.com/Ranzz02/uno-discord-bot/src/bot"
	"github.com/Ranzz02/uno-discord-bot/src/config"
)

func main() {
	// Check configs
	config.NewConf()

	// Init bot
	bot.InitBot()
}
