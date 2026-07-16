package redis

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() (*redis.Client, error) {
	addr := os.Getenv("LIGHTCHAT_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	dbIndex := 0
	if rawDB := os.Getenv("LIGHTCHAT_REDIS_DB"); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			return nil, err
		}
		dbIndex = parsedDB
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("LIGHTCHAT_REDIS_PASSWORD"),
		DB:           dbIndex,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
