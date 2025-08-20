package service

import (
	"fmt"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"github.com/hbrawnak/go-linko/internal/utils"
	"log"
	"time"
)

type Service struct {
	Models data.Models
	Redis  database.RedisClient
}

func (s *Service) GenerateShortCode() string {
	code := s.GetShortCode()
	return utils.HashToBase62(code)
}

func (s *Service) GetShortCode() string {
	incr := s.Redis.INCR()
	padded := fmt.Sprintf("%0*d", utils.ShortCodeLenMin, incr)
	return padded
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

func (s *Service) StoreInRedisCacheBG(key string, values map[string]string) {
	go func() {
		if err := s.Redis.HSet(key, values); err != nil {
			log.Printf("failed to store in redis cache: %v", err)
		}
	}()
}
