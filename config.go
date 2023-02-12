package main

import (
	"github.com/caarlos0/env/v7"
	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
)

const envPrefix = "CAPTCHA_THE_BOT_"

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN" validate:"required"`
	WebhookBase   string `env:"WEBHOOK_BASE"   validate:"required,url"`
	WebhookPath   string `env:"WEBHOOK_PATH"   validate:"required,uri"`
	ListenAddress string `env:"LISTEN_ADDRESS" validate:"required,hostname_port"`
	DebugMode     bool   `env:"DEBUG_MODE"     validate:"-"`
}

func LoadConfig() Config {
	cfg := Config{}
	err := env.Parse(&cfg, env.Options{
		Prefix: envPrefix,
	})
	assert(err == nil, "Reading config:", err)

	validate := validator.New()
	err = validate.Struct(cfg)
	assert(err == nil, "Invalid config:", err)

	return cfg
}
