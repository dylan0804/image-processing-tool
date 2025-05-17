package storage

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/dylan0804/image-processing-tool/internal/api/interfaces"
	"github.com/redis/go-redis/v9"
)

type RedisSessionStore interface {
	Set(ctx context.Context, sessionID string, data interfaces.SessionData) error
	Get(ctx context.Context, sessionID string) (interfaces.SessionData, bool, error)
	Delete(ctx context.Context, sessionID string) error
}

type RedisSessionImpl struct {
	Client *redis.Client
	Ttl time.Duration
}

func NewRedisSessionStore() (*RedisSessionImpl, error) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")

	if host == "" {
		log.Printf("REDIS_HOST not set, defaulting to localhost")
		host = "localhost"
	}
	if port == "" {
		log.Printf("REDIS_PORT not set, defaulting to 6379")
		port = "6379"
	}
	
	redisAddr := host + ":" + port // Construct address from env vars
	log.Printf("Attempting to connect to Redis at: %s", redisAddr)

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr, // Use the constructed address
		Password: "",
		DB: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("WARNING: Redis connection failed: %v", err)
		return nil, err
	}

	return &RedisSessionImpl{
		Client: rdb,
		Ttl: 30*time.Minute,
	}, nil
}

func (r *RedisSessionImpl) Set(ctx context.Context, sessionID string, data interfaces.SessionData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, "session:"+sessionID, jsonData, r.Ttl).Err()
}

func (r *RedisSessionImpl) Get(ctx context.Context, sessionID string) (interfaces.SessionData, bool, error) {
	result, err := r.Client.Get(ctx, "session:"+sessionID).Result()
	if err == redis.Nil {
		return interfaces.SessionData{}, false, nil
	} else if err != nil {
		return interfaces.SessionData{}, false, err
	}

	var data interfaces.SessionData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return interfaces.SessionData{}, false, err
	}

	return data, true, nil
}

func (r *RedisSessionImpl) Delete(ctx context.Context, sessionID string) error {
	return r.Client.Del(ctx, "session:"+sessionID).Err()
}