package redis_config

import (
	"context"
	"fmt"
	"log"
	"movie-ticket/config"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	RedisClient *redis.Client
)

func InitRedis() {
	redisHost := config.Get("REDIS_HOST")
	redisPort := config.Get("REDIS_PORT")

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(ctx).Result()

	if err != nil {
		log.Fatalf("Tidak bisa konek ke Redis: %v", err)
	}

	log.Println("✅ Redis terkoneksi")
}
