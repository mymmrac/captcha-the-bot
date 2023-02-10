package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := LoadConfig()

	bot, err := telego.NewBot(cfg.TelegramToken, telego.WithDefaultDebugLogger())
	assert(err == nil, "Create bot:", err)

	secretTokenData, err := bcrypt.GenerateFromPassword([]byte(cfg.TelegramToken), bcrypt.DefaultCost)
	assert(err == nil, "Generate secret:", err)
	secretToken := fmt.Sprintf("%X", secretTokenData)

	updates, err := bot.UpdatesViaWebhook(
		cfg.WebhookPath,
		telego.WithWebhookServer(telego.FastHTTPWebhookServer{
			Logger:      bot.Logger(),
			Server:      &fasthttp.Server{},
			Router:      router.New(),
			SecretToken: secretToken,
		}),
		telego.WithWebhookSet(&telego.SetWebhookParams{
			URL:         cfg.WebhookBase + cfg.WebhookPath,
			SecretToken: secretToken,
		}),
	)
	assert(err == nil, "Get updates", err)

	bh, err := th.NewBotHandler(bot, updates, th.WithStopTimeout(time.Second*10))
	assert(err == nil, "Setup bot handler", err)

	handler := NewHandler(bot, bh)
	handler.Init()
	handler.RegisterHandlers()

	done := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("Stopping...")

		err = bot.StopWebhook()
		if err != nil {
			fmt.Println("ERROR: Stop webhook:", err)
		}

		bh.Stop()

		done <- struct{}{}
	}()

	go bh.Start()

	go func() {
		fmt.Println("Listening for updates...")
		err = bot.StartWebhook(cfg.ListenAddress)
		assert(err == nil, "Start webhook:", err)
	}()

	<-done
	fmt.Println("Done")
}

func assert(ok bool, args ...any) {
	if !ok {
		fmt.Println(append([]any{"FATAL:"}, args...)...)
		os.Exit(1)
	}
}
