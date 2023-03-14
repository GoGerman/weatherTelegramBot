package main

import (
	"github.com/GoGerman/weatherTelegramBot/telegram"
	"log"
)

func main() {
	bot, err := telegram.NewBot("<6225684885:AAHhd4JF6cIO1eEZ9Vo5jaQvB4A_z9bIQbE\n>")
	if err != nil {
		log.Fatal(err)
	}

	bot.Start()
}
