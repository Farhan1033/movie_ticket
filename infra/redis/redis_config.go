package redis_config

import (
	"context"
	"fmt"
	"log"
	"movie-ticket/config"
	"net/url"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	RedisClient *redis.Client
)

func InitRedis() {
	redisHost := config.Get("REDIS_HOST")
	redisPort := config.Get("REDIS_PORT")
	redisPass := config.Get("REDIS_PASS")
	encodedPass := url.QueryEscape(redisPass)

	redisURL := fmt.Sprintf("rediss://default:%s@%s:%s", encodedPass, redisHost, redisPort)

	//rediss://default:Ae00000MMXhxTu+/j7ZQFB9x7ol7NPpaVTEDrnTDnpluBbE41wQw8oQ5h+RsbbCzNjqZXRk@movie_ticket-kxcp-cbdb-053533.leapcell.cloud:6379

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Gagal parse Redis URL: %v", err)
	}

	RedisClient = redis.NewClient(opt)

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Tidak bisa konek ke Redis: %v", err)
	}

	log.Println("✅ Redis terkoneksi")
}
