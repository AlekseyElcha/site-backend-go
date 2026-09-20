package redis_client

import (
	"context"
	"fmt"
	"time" // Добавьте импорт пакета time

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisClient(rdb *redis.Client) *RedisClient {
	return &RedisClient{
		rdb: rdb,
	}
}

func (s *RedisClient) AddAuthCodeToRedis(ctx context.Context, requesterEmail string, authCode string) error {
	key := fmt.Sprintf("auth_code_req:%s", requesterEmail)

	if err := s.rdb.Set(ctx, key, authCode, 15*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to save auth code for %s to redis: %w", requesterEmail, err)
	}

	return nil
}

func (s *RedisClient) IsAuthCodeFromUserValid(ctx context.Context, requesterEmail string, authCode string) (bool, error) {
	key := fmt.Sprintf("auth_code_req:%s", requesterEmail)
	// fmt.Println(key)
	value, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to get auth code for %s from redis: %w", requesterEmail, err)
	}
	if value != authCode {
		return false, nil
	}

	s.rdb.Del(ctx, key)

	return true, nil
}
