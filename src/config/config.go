package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

var Conf *Config

type Config struct {
	Token string `env:"DISCORD_TOKEN,required"`
}

func NewConf() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Unable to load env: %v", err)
	}

	config := &Config{}

	if err := env.Parse(config); err != nil {
		log.Fatalf("Unable to load variables from env: %v", err)
	}

	Conf = config
}
