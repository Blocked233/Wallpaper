package cache

import (
	"github.com/go-redis/redis"
)

var (
	RedisClient *redis.Client
)

func NewRedisClient(redisPassword string) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: redisPassword,
		DB:       0,
	})
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
