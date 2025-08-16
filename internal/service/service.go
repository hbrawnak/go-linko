package service

import (
	"crypto/rand"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"log"
	"regexp"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var base62Regex = regexp.MustCompile("^[a-zA-Z0-9]+$")

const shortCodeLenMin = 7
const shortCodeLenMax = 10
const redisTTL = 24 * time.Hour

type Service struct {
	Models data.Models
	Redis  database.RedisClient
}

func (s *Service) GenerateShotCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := 0; i < n; i++ {
		b[i] = letters[int(b[i])%len(letters)]
	}

	return string(b), nil
}

func (s *Service) IsBase62(code string) bool {
	return base62Regex.MatchString(code)
}

func (s *Service) IsLengthOk(code string) bool {
	return len(code) >= shortCodeLenMin && len(code) <= shortCodeLenMax
}

func (s *Service) UpdateHitCountBG(c string) {
	go func(c string) {
		const maxRetries = 3
		const retryDelay = 200 * time.Millisecond

		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := s.Models.URL.IncrementHitCount(c)
			if err == nil {
				return
			}

			log.Printf("failed to update hit count (attempt %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
			}
		}
	}(c)
}

func (s *Service) StoreInRedisCacheBG(key string, value string) {
	go func() {
		if err := s.Redis.Set(key, value, redisTTL); err != nil {
			log.Printf("failed to store in redis cache: %v", err)
		}
	}()
}
