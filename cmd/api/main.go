package main

import (
	"database/sql"
	"fmt"
	"github.com/hbrawnak/go-linko/data"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const port = "8080"

var count int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Printf("URL shortener service on port %s\n", port)

	db := connectToDB()
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

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		db, err := openDB(dsn)

		if err != nil {
			log.Println("Error opening database connection")
			count++
		} else {
			log.Println("Database connected!!")
			return db
		}

		if count > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for 2 seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
