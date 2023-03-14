package telegram

import (
	"log"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	//"github.com/myusername/go-telegram-bot/postgres"
	//"github.com/myusername/go-telegram-bot/storage"
)

type Bot struct {
	api   *tgbotapi.BotAPI
	store *storage.Storage
}

func NewBot(apiToken, dbURL string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bot")
	}

	db, err := postgres.NewConnection(dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	store := storage.NewStorage(db)

	return &Bot{api: api, store: store}, nil
}

func (bot *Bot) Start() error {
	log.Println("Bot has started")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(u)
	if err != nil {
		return errors.Wrap(err, "failed to get updates channel")
	}

	for update := range updates {
		if update.Message == nil { // игнорируем не-сообщения
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		switch update.Message.Text {
		case "Информация":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пока недоступно")
			bot.api.Send(msg)

		case "Статистика":
			requests, err := bot.store.GetRequestsByUserID(update.Message.From.ID)
			if err != nil {
				log.Println(err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при получении статистики")
				bot.api.Send(msg)
				continue
			}

			if len(requests) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас ещё не было запросов")
				bot.api.Send(msg)
				continue
			}

			firstRequest := requests[0]
			totalRequests := len(requests)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Статистика:")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Первый запрос: "+firstRequest.RequestDate.Format(time.RFC822)),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Всего запросов: "+strconv.Itoa(totalRequests)),
				),
			)
			bot.api.Send(msg)

		default:
			// обработка неизвестной команды
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды")
			bot.api.Send(msg)
		}

		// сохраняем запрос в базу данных
		request := storage.Request{
			UserID:      update.Message.From.ID,
			RequestType: update.Message.Text,
			RequestDate: time.Now(),
		}

		err := bot.store.SaveRequest(&request)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
