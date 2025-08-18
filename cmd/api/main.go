package main

import (
	"database/sql"
	"fmt"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"github.com/hbrawnak/go-linko/internal/handlers"
	"github.com/hbrawnak/go-linko/internal/routes"
	"github.com/hbrawnak/go-linko/internal/service"
	"github.com/hbrawnak/go-linko/internal/worker"
	"log"
	"net/http"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const port = "8080"

type Config struct {
	DB      *sql.DB
	Service *service.Service
	Queue   chan worker.URLTask
}

func NewConfig() *Config {
	db := database.ConnectToDB()
	if db == nil {
		log.Panic("Failed to connect to database")
	}

	redisClient := database.ConnectToRedis()
	if redisClient == nil {
		log.Panic("Failed to connect to Redis")
	}

	models := data.New(db)

	svc := &service.Service{
		Models: models,
		Redis:  *redisClient,
	}

	// Create task queue channel
	taskQueue := make(chan worker.URLTask, 10)

	return &Config{
		DB:      db,
		Service: svc,
		Queue:   taskQueue,
	}
}

func main() {
	log.Printf("URL shortener service on port %s\n", port)

	app := NewConfig()

	// Start worker goroutine to process tasks from queue
	go worker.StartURLTaskWorker(app.Queue, app.Service)

	// Create handler with service dependency
	handler := handlers.NewHandler(app.Service, app.Queue)

	svr := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: routes.SetupRoutes(handler),
	}

	// Starting server
	if err := svr.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
