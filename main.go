package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type dbConfig struct {
	user     string
	password string
	host     string
	port     string
	dbName   string
	dbType   string
}

func main() {
	dbCfg := dbConfig{
		user:     "user",
		password: "password",
		host:     "localhost",
		port:     "3306",
		dbName:   "dbApisUser",
		dbType:   "sqlite3",
	}

	// Подключаемся к базе данных
	db, err := dbConnect(dbCfg)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
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

		userID := update.Message.From.ID
		username := update.Message.From.UserName
		command := update.Message.Command()

		if command == "" {
			continue
		}

		err := saveRequest(userID, username, command)
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
	userID := update.Message.From.ID
	username := update.Message.From.UserName

	_ = username

	var totalRequests int
	var firstRequestTime time.Time

	row := db.QueryRow("SELECT COUNT(*), MIN(request_time) FROM users WHERE id = ?", userID)
	err := row.Scan(&totalRequests, &firstRequestTime)
	if err != nil {
		log.Printf("Unable to get user requests: %v\n", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Total requests: %d\nFirst request time: %s", totalRequests, firstRequestTime.Format("2006-01-02 15:04:05")))
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}

func saveRequest(userID int, username string, command string) error {
	_, err := db.Exec("INSERT INTO users(id, username, command, request_time) VALUES (?, ?, ?, ?)", userID, username, command, time.Now())
	if err != nil {
		return fmt.Errorf("unable to save request: %v", err)
	}
	return nil
}

func getWeather(city string) (string, error) {
	apiKey := "78cfb04016855233daaf20a3817aefa3"
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, apiKey)

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

func dbConnect(config dbConfig) (*sql.DB, error) {
	db, err := sql.Open(config.dbType, config.dbName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Создание базы данных
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s", config.dbName))
	if err != nil {
		return nil, err
	}

	// Выбор базы данных
	_, err = db.Exec(fmt.Sprintf("USE %s", config.dbName))
	if err != nil {
		return nil, err
	}

	// Создание таблицы
	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INT NOT NULL AUTO_INCREMENT,
            PRIMARY KEY (id),
            username varchar(255), 
            command varchar(255),
            request_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        );
    `)
	if err != nil {
		return err
	}
	return nil
}
