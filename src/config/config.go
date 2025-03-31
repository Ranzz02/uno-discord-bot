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
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: .env file not found. Relying on environment variables instead.")
	}

	config := &Config{}

	if err := env.Parse(config); err != nil {
		log.Fatalf("Unable to load variables from env: %v", err)
	}

	Conf = config
}
