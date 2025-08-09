package redis_config

import (
	"context"
	"log"
	"movie-ticket/config"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	RedisClient *redis.Client
)

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Get("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(ctx).Result()

	if err != nil {
		log.Fatalf("Tidak bisa konek ke Redis: %v", err)
	}

	log.Println("✅ Redis terkoneksi")
}
