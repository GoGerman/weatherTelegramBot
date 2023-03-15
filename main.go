package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

type dbConfig struct{ user, password, host, port, dbName string }

var db *sql.DB

func main() {
	dbCfg := dbConfig{user: "user", password: "password", host: "localhost", port: "3306", dbName: "111"}
	var err error
	db, err = sql.Open("sqlite3", dbCfg.dbName)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	err = createTables(db)
	if err != nil {
		log.Fatalf("Cannot create database: %v\n", err)
	}
	defer db.Close()

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

		userID, command := update.Message.From.ID, update.Message.Command()

		if command == "" {
			continue
		}

		if err := saveRequest(userID, update.Message.From.UserName, command); err != nil {
			log.Printf("Unable to save request: %v\n", err)
		}

		switch command {
		case "info":
			handleInfoCommand(bot, &update)
		case "stat":
			handleStatCommand(bot, &update)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
			if _, err := bot.Send(msg); err != nil {
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
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}

func handleStatCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userID := update.Message.From.ID
	var totalRequests int
	var firstRequestTime time.Time

	err := db.QueryRow("SELECT COUNT(*), MIN(CAST(request_time AS DATETIME)) FROM users WHERE id = ?", userID).
		Scan(&totalRequests, &firstRequestTime)

	if err != nil {
		log.Printf("Unable to get user requests: %v\n", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("Total requests: %d\nFirst request time: %v",
			totalRequests, firstRequestTime))

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}

func saveRequest(userID int, username, command string) error {
	stmt, err := db.Prepare(`
       INSERT INTO users (id, username, request_time, command)
       VALUES (:id, :username, :request_time, :command)
   `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sql.Named("id", userID), sql.Named("username", username), sql.Named("request_time", time.Now()), sql.Named("command", command))
	if err != nil {
		return err
	}

	return nil
}

func getWeather(city string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s",
		city, "78cfb04016855233daaf20a3817aefa3"))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var weatherData struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return "", err
	}

	return fmt.Sprintf("Temperature: %.1f°C\nHumidity: %d%%\nWind Speed: %.1f m/s",
		weatherData.Main.Temp-273.15, weatherData.Main.Humidity, weatherData.Wind.Speed), nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
		    id INTEGER PRIMARY KEY,
			username TEXT,
			command TEXT,
			request_time TIMESTAMP
		);
	`)
	return err
}
