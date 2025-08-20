package service

import (
	"encoding/json"
	"errors"
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

type StatsData struct {
	Code        string `json:"code"`
	Count       int64  `json:"count"`
	LastAccess  string `json:"update_at,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
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

func (s *Service) GetStats(code string) (*StatsData, error) {
	cacheKey := "stats:" + code

	if cached, err := s.Redis.Get(cacheKey); err == nil && cached != "" {
		var stats StatsData
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return &stats, nil
		}
		// If unmarshal fails, fallback to DB
	}

	// DB lookup if cache miss
	u, err := s.Models.URL.GetOne(code)
	if err != nil {
		return nil, errors.New("short code not found")
	}

	stats := &StatsData{
		Code:        u.ShortCode,
		Count:       u.HitCount,
		LastAccess:  u.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		OriginalURL: u.OriginalURL,
	}

	// Cache result for next time
	if jsonData, err := json.Marshal(stats); err == nil {
		if err := s.Redis.Set(cacheKey, string(jsonData)); err != nil {
			log.Println("failed to cache stats: ", err)
		}
	}

	return stats, nil
}
