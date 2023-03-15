package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	bot, err := tgbotapi.NewBotAPI("6225684885:AAHhd4JF6cIO1eEZ9Vo5jaQvB4A_z9bIQbE")
	if err != nil {
		log.Fatalf("Unable to connect to Telegram API: %v\n", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Command() {
		case "info":
			handleInfoCommand(bot, &update)
		case "stat":
			handleStatCommand(bot, &update)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message: %v\n", err)
			}
		}
	}
}

func handleInfoCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	city := update.Message.CommandArguments()

	weather, err := getWeather(city)
	if err != nil {
		log.Printf("Can't get weather: %v\n", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, weather)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}

func handleStatCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	var totalRequests int
	var firstRequestTime time.Time

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Total requests: %d\nFirst request time: %s", totalRequests, firstRequestTime.Format("2006-01-02 15:04:05")))
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}

func getWeather(city string) (string, error) {
	apiKey := "78cfb04016855233daaf20a3817aefa3"
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var weatherData struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return "", err
	}

	temperature := strconv.FormatFloat(weatherData.Main.Temp-273.15, 'f', 1, 64)
	humidity := strconv.Itoa(weatherData.Main.Humidity)
	windSpeed := strconv.FormatFloat(weatherData.Wind.Speed, 'f', 1, 64)

	return fmt.Sprintf("Temperature: %s°C\nHumidity: %s%%\nWind Speed: %s m/s", temperature, humidity, windSpeed), nil
}
