package main

import (
	"log"
	"weatherTelegramBot/weatherTelegramBot/telegram"
)

func main() {
	bot, err := telegram.NewBot("<6225684885:AAHhd4JF6cIO1eEZ9Vo5jaQvB4A_z9bIQbE")
	if err != nil {
		log.Fatal(err)
	}

	bot.Start()
}
