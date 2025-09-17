package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zavgorodniyvv/GymBot/internal/bot"

	"log"
)

func main() {
	token := "******"
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	api.Debug = false
	log.Printf("Authorized on account %s", api.Self.UserName)
	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60
	updates := api.GetUpdatesChan(upd)
	b := bot.New(api)
	for update := range updates {
		b.Handle(update)
	}

}
