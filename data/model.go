package data

import (
	"context"
	"database/sql"
	"time"
)

const dbTimeout = time.Second * 3

var db *sql.DB

func New(dbPool *sql.DB) Models {
	db = dbPool
	return Models{
		URL: URL{},
	}
}

type Models struct {
	URL URL
}

type URL struct {
	ID          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *URL) Insert(url URL) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	stmt := `insert into urls (short_code, original_url, created_at, updated_at)
		values ($1, $2, $3, $4) returning id`

	err := db.QueryRowContext(ctx, stmt,
		url.ShortCode,
		url.OriginalURL,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}
