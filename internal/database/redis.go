package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

var RedisRetryCount int64

const IncrKey = "url_counter"

const dbTimeout = time.Second * 3
const defaultTTLRedis = 24 * time.Hour

type CachedURL struct {
	URL       string `json:"url"`
	Persisted string `json:"persisted"`
}

func (c CachedURL) ToMap() map[string]string {
	return map[string]string{
		"url":       c.URL,
		"persisted": c.Persisted,
	}
}

func ConnectToRedis() *RedisClient {
	redisURL := os.Getenv("REDIS_DSN")
	var client *redis.Client

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Unable to parse Redis URL: %s", err)
	}

	for {
		client = redis.NewClient(options)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := client.Ping(ctx).Err()
		if err != nil {
			log.Println("Failed to connect to Redis, retrying...", err)
			RedisRetryCount++
		} else {
			log.Println("Redis connected successfully!")
			return &RedisClient{client: client}
		}

		if RedisRetryCount > 10 {
			log.Fatalf("Could not connect to Redis after %d attempts: %v", RedisRetryCount, err)
		}

		log.Println("Backing off for 2 seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}

func (r *RedisClient) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	log.Println("Getting cache")
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(key string, value string, ttl ...time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	expire := defaultTTLRedis
	if len(ttl) > 0 {
		expire = ttl[0]
	}

	log.Println("Setting cache")
	return r.client.Set(ctx, key, value, expire).Err()
}

func (r *RedisClient) HSet(key string, values map[string]string, ttl ...time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	err := r.client.HSet(ctx, key, values).Err()
	if err != nil {
		log.Printf("[Redis:HSet] failed to set cache for key=%q field: %v", key, err)
		return err
	}

	expire := defaultTTLRedis
	if len(ttl) > 0 {
		expire = ttl[0]
	}
	return r.client.Expire(ctx, key, expire).Err()
}

func (r *RedisClient) HGet(key, field string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	val, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("[Redis:HGet] key=%q field=%q not found", key, field)
		} else {
			log.Printf("[Redis:HGet] failed for key=%q field=%q: %v", key, field, err)
		}
		return "", err
	}

	return val, nil
}

func (r *RedisClient) INCR() int64 {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	incr, err := r.client.Incr(ctx, IncrKey).Result()
	if err != nil {
		log.Println("Failed to incr url_counter")
	}

	log.Println("Successfully incr url_counter", incr)
	return incr
}
