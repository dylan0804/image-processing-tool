package storage

import (
	"context"
	"encoding/json"
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
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
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
	result, err := r.Client.Get(ctx, "session"+sessionID).Result()
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