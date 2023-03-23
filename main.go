package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
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
		log.Printf("Cannot create database: %v\n", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			return
		}
	}()

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

		var totalRequests int
		err = db.QueryRow("select totalRequests from users where id = ?", userID).Scan(&totalRequests)

		err = saveRequest(userID, update.Message.From.UserName, command, totalRequests)
		if err != nil {
			log.Printf("Unable to save request: %v\n", err)
		}

		switch command {
		case "info":
			handleInfoCommand(bot, &update)
		case "stat":
			handleStatCommand(bot, &update)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
			if _, err = bot.Send(msg); err != nil {
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
	if _, err = bot.Send(msg); err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}

func handleStatCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userID := update.Message.From.ID
	var totalRequests int
	var firstRequestTime string

	err := db.QueryRow("select totalRequests, min(request_time) from users where id = ?", userID).
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

func saveRequest(userID int, username, command string, totalRequests int) error {
	var count int
	errCount := db.QueryRow("select count(*) from users where id = ?", userID).Scan(&count)
	if errCount != nil {
		return errCount
	}
	if count > 0 {
		return nil
	}

	if totalRequests == 0 {
		stmt, err := db.Prepare(`
        insert into users (id, username, request_time, command, totalRequests)
        values (?, ?, ?, ?, ?)
        `)
		if err != nil {
			return err
		}

		defer func() {
			err = stmt.Close()
			if err != nil {
				return
			}
		}()

		_, err = stmt.Exec()
		if err != nil {
			return err
		}

		_, err = stmt.Exec(userID, username, time.Now(), command, 1)
		if err != nil {
			return err
		}
	} else {
		stmt, err := db.Prepare(`
		update users
		set totalRequests = ?
		where id = ?
		`)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(totalRequests+1, userID)
		if err != nil {
			return err
		}

		defer func() {
			err = stmt.Close()
			if err != nil {
				return
			}
		}()
	}

	return nil
}

func getWeather(city string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s",
		city, "78cfb04016855233daaf20a3817aefa3"))
	if err != nil {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()

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

	return fmt.Sprintf("Temperature (Температура): %.1f°C\nHumidity (Видимость): %d%%\nWind Speed (Скорость ветра): %.1f m/s",
		weatherData.Main.Temp-273.15, weatherData.Main.Humidity, weatherData.Wind.Speed), nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		create table users (
		    id integer primary key,
			username TEXT NOT NULL,
			command TEXT NOT NULL,
			request_time timestamp,
		    totalRequests integer default 0
		);
	`)
	return err
}
