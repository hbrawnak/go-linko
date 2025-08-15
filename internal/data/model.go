package data

import (
	"context"
	"database/sql"
	"log"
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
	HitCount    int64     `json:"hit_count"`
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

func (u *URL) GetOne(code string) (*URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := "select id, short_code, original_url, created_at, updated_at from urls where short_code = $1"

	var url URL

	row := db.QueryRowContext(ctx, query, code)
	err := row.Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.CreatedAt, &url.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (u *URL) IncrementHitCount(c string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	log.Printf("Increment hit count for code: %s\n", c)

	query := `
		UPDATE urls
		SET hit_count = hit_count + 1, updated_at = NOW()
		WHERE short_code = $1
	`

	_, err := db.ExecContext(ctx, query, c)

	if err != nil {
		log.Printf("Error increment hit count for %s. %s\n\n", c, err)
		return err
	}

	return nil
}
