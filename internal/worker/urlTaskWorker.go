package worker

import (
	"fmt"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"github.com/hbrawnak/go-linko/internal/service"
	"log"
	"time"
)

type URLTask struct {
	ShortCode   string
	OriginalURL string
}

func StartURLTaskWorker(taskQueue <-chan URLTask, service *service.Service) {
	log.Println("Worker started and listening for tasks...")

	go func() {
		for task := range taskQueue {
			log.Printf("Processing task: Code=%s, URL=%s", task.ShortCode, task.OriginalURL)
			if err := processURLTask(task, service); err != nil {
				log.Printf("Error processing task %s: %v", task.ShortCode, err)
			} else {
				log.Printf("Successfully processed task: %s", task.ShortCode)
			}
		}
	}()
}

func processURLTask(task URLTask, service *service.Service) error {
	const maxRetries = 3
	retryDelay := 200 * time.Millisecond

	u := data.URL{
		ShortCode:   task.ShortCode,
		OriginalURL: task.OriginalURL,
	}

	fields := database.CachedURL{
		URL:       task.OriginalURL,
		Persisted: "1",
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := service.Models.URL.Insert(u)
		if err == nil {
			log.Printf("Insert succeeded in db for shortcode=%s on attempt %d", u.ShortCode, attempt)
			err := service.Redis.HSet(task.ShortCode, fields.ToMap())
			if err != nil {
				return err
			}
			log.Printf("Data persisted successfully for both db and redis %s", u.ShortCode)
			return nil
		}

		lastErr = err
		log.Printf("Insert failed for shortcode=%s (attempt %d/%d): %v", u.ShortCode, attempt, maxRetries, err)

		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return fmt.Errorf("failed to insert URL (shortcode=%s, url=%s) after %d attempts: %w",
		u.ShortCode, u.OriginalURL, maxRetries, lastErr)
}
