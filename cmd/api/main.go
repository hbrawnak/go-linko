package main

import (
	"database/sql"
	"fmt"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"log"
	"net/http"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const port = "8080"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Printf("URL shortener service on port %s\n", port)

	db := database.ConnectToDB()
	if db == nil {
		log.Panic("Failed to connect to database")
	}

	app := &Config{
		DB:     db,
		Models: data.New(db),
	}

	svr := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.routes(),
	}

	// Starting server
	err := svr.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
