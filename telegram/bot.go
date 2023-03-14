package telegram

import (
	"database/sql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Bot struct {
	botAPI     *tgbotapi.BotAPI
	db         *sql.DB
	dbHost     string
	dbPort     string
	dbUser     string
	dbPassword string
	dbName     string
	dbSSLMode  string
}

func NewBot(token string, dbHost string, dbPort string, dbUser string, dbPassword string, dbName string, dbSSLMode string) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	// подключение к БД
	db, err := sql.Open("postgres", getDBConnString(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Bot{
		botAPI:     botAPI,
		db:         db,
		dbHost:     dbHost,
		dbPort:     dbPort,
		dbUser:     dbUser,
		dbPassword: dbPassword,
		dbName:     dbName,
		dbSSLMode:  dbSSLMode,
	}, nil
}

func (b *Bot) Run() error {
	// установка обработчика сообщений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот, который может помочь тебе найти информацию о погоде и показать статистику твоих запросов. Чтобы узнать, что я умею, отправь мне /help")
			b.botAPI.Send(msg)
		case "/help":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я поддерживаю две команды:\n- /weather [город] - поиск информации о погоде в заданном городе\n- /stats - показать статистику твоих запросов")
			b.botAPI.Send(msg)
		case "/weather":
			// получение города из запроса
			city := update.Message.CommandArguments()

			// отправка запроса на сайт для получения погоды
			weatherInfo, err := getWeatherInfo(city)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при поиске информации о погоде")
				b.botAPI.Send(msg)
				continue
			}

			// сохранение запроса в БД
			if err := saveRequest(b.db, update.Message.Chat.ID, "weather", city); err != nil {
				log.Printf("Error saving request to DB: %v", err)
			}

			// отправка информации о погоде
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, weatherInfo)
			b.botAPI.Send(msg)
		case "/stats":
			// получение статистики запросов
			stats, err := getRequestStats(b.db, update.Message.Chat.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении статистики")
				b.botAPI.Send(msg)
				continue
			}

			// формирование ответа с статистикой
			response := "Статистика:\n"
			response += "Первый запрос: " + stats.FirstRequestTime.Format("2006-01-02 15:04:05") + "\n"
			response += "Количество запросов: " + stats.NumRequests + "\n"

			// отправка ответа
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			b.botAPI.Send(msg)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не понимаю, что ты хочешь. Попробуй отправить /help")
			b.botAPI.Send(msg)
		}
	}

	return nil
}

// закрытие соединения с БД
func (b *Bot) Close() error {
	return b.db.Close()
}

// формирование строки подключения к БД
func getDBConnString(dbHost string, dbPort string, dbUser string, dbPassword string, dbName string, dbSSLMode string) string {
	connString := "host=" + dbHost +
		" port=" + dbPort +
		" user=" + dbUser +
		" password=" + dbPassword +
		" dbname=" + dbName +
		" sslmode=" + dbSSLMode
	return connString
}

// получение информации о погоде с сайта
func getWeatherInfo(city string) (string, error) {
	// код для получения информации о погоде
	return "Информация о погоде в городе " + city, nil
}

// сохранение запроса в БД
func saveRequest(db *sql.DB, chatID int64, command string, args string) error {
	// код для сохранения запроса в БД
	return nil
}

// получение статистики запросов из БД
func getRequestStats(db *sql.DB, chatID int64) (*RequestStats, error) {
	// код для получения статистики запросов из БД
	return &RequestStats{
		FirstRequestTime: time.Now(),
		NumRequests:      "0",
	}, nil
}
