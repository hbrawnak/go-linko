package main

import (
	"database/sql"
	"fmt"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"github.com/hbrawnak/go-linko/internal/service"
	"log"
	"net/http"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const port = "8080"

type Config struct {
	DB      *sql.DB
	Models  data.Models
	Redis   *database.RedisClient
	Service service.Service
}

func main() {
	log.Printf("URL shortener service on port %s\n", port)

	db := database.ConnectToDB()
	if db == nil {
		log.Panic("Failed to connect to database")
	}

	connectToRedis := database.ConnectToRedis()
	if connectToRedis == nil {
		log.Panic("Failed to connect to connectToRedis")
	}

	app := &Config{
		DB:     db,
		Models: data.New(db),
		Redis:  connectToRedis,
		Service: service.Service{
			Models: data.New(db),
			Redis:  *connectToRedis,
		},
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
