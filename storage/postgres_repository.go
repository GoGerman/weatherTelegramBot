package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"weatherTelegramBot/weatherTelegramBot/telegram"
)

// PostgresRepo is a PostgreSQL implementation of the storage.Repository interface
type PostgresRepo struct {
	db *pgx.Conn
}

// NewPostgresRepo returns a new instance of PostgresRepo
func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	return &PostgresRepo{db: conn}, nil
}

// Close closes the database connection
func (repo *PostgresRepo) Close() {
	repo.db.Close(context.Background())
}

// SaveRequest saves a new request in the database
func (repo *PostgresRepo) SaveRequest(request *telegram.Request) error {
	query := `
		INSERT INTO requests (chat_id, command, args, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	var id int64
	err := repo.db.QueryRow(
		context.Background(),
		query,
		request.ChatID,
		request.Command,
		request.Args,
		request.CreatedAt,
	).Scan(&id)
	if err != nil {
		return errors.Wrap(err, "failed to save request")
	}

	request.ID = id
	return nil
}

// GetRequestStats returns the request statistics for a given chat
func (repo *PostgresRepo) GetRequestStats(chatID int64) (*telegram.RequestStats, error) {
	query := `
		SELECT created_at, COUNT(*)
		FROM requests
		WHERE chat_id = $1
		GROUP BY chat_id;
	`
	var (
		firstRequestTime time.Time
		numRequests      int
	)
	err := repo.db.QueryRow(
		context.Background(),
		query,
		chatID,
	).Scan(&firstRequestTime, &numRequests)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get request stats")
	}

	stats := &telegram.RequestStats{
		FirstRequestTime: firstRequestTime,
		NumRequests:      fmt.Sprintf("%d", numRequests),
	}

	return stats, nil
}
