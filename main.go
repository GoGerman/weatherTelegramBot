package main

import (
	"github.com/GoGerman/weatherTelegramBot/telegram"
	"log"
)

func main() {
	bot, err := telegram.NewBot("<your_bot_token>")
	if err != nil {
		log.Fatal(err)
	}

	bot.Start()
}
