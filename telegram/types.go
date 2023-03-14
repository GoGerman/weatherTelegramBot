package telegram

import "time"

// структура для хранения информации о запросе
type Request struct {
	ID        int64
	ChatID    int64
	Command   string
	Args      string
	CreatedAt time.Time
}

// структура для хранения информации о статистике запросов
type RequestStats struct {
	FirstRequestTime time.Time
	NumRequests      string
}
