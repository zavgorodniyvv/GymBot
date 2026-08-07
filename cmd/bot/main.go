package main

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zavgorodniyvv/GymBot/internal/bot"
	"github.com/zavgorodniyvv/GymBot/internal/logging"
	"github.com/zavgorodniyvv/GymBot/internal/storage"

	"log"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	uri := os.Getenv("MONGO_URI")

	// Ставим фильтр до первого лога: сетевые ошибки печатают URL Telegram API
	// вместе с токеном, в том числе из горутины GetUpdatesChan.
	secrets := append([]string{token}, logging.SecretsFromURI(uri)...)
	log.SetOutput(logging.NewRedactingWriter(os.Stderr, secrets...))

	if token == "" {
		log.Fatal("TELEGRAM_TOKEN is empty")
	}
	if uri == "" {
		log.Fatal("MONGO_URI is empty")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	// Подключаемся к MongoDB
	st, err := storage.NewMongoStorage(uri)
	if err != nil {
		log.Fatal("Mongo connection error:", err)
	}

	b := bot.New(api, st)

	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60
	updates := api.GetUpdatesChan(upd)

	for update := range updates {
		b.Handle(update)
	}
}
