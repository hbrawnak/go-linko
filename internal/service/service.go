package service

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/hbrawnak/go-linko/internal/data"
	"github.com/hbrawnak/go-linko/internal/database"
	"log"
	"regexp"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var base62Regex = regexp.MustCompile("^[a-zA-Z0-9]+$")

const shortCodeLenMin = 7
const shortCodeLenMax = 8
const redisTTL = 24 * time.Hour

type Service struct {
	Models data.Models
	Redis  database.RedisClient
}

func (s *Service) GenerateShortCode() string {
	code := s.GetShortCode()
	return hashToBase62(code)
}

func (s *Service) GetShortCode() string {
	incr := s.Redis.INCR()
	padded := fmt.Sprintf("%0*d", shortCodeLenMin, incr)
	return padded
}

func ToBase62(num uint64) string {
	if num == 0 {
		return string(letters[0])
	}
	var encoded []byte
	for num > 0 {
		remainder := num % 62
		encoded = append([]byte{letters[remainder]}, encoded...)
		num = num / 62
	}
	return string(encoded)
}

func hashToBase62(input string) string {
	hash := sha256.Sum256([]byte(input))

	num := binary.BigEndian.Uint64(hash[:8])
	code := ToBase62(num)

	for len(code) < shortCodeLenMin {
		code = "0" + code
	}

	if len(code) > shortCodeLenMax {
		code = code[:shortCodeLenMax]
	}

	return code
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
