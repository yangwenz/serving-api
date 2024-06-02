package utils

import (
	"context"
	goredis "github.com/redis/go-redis/v9"
)

func NewRedisClient(redisAddress string) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0, // use default DB
	})
	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}
