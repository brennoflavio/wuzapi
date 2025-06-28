package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         context.Context
)

func InitRedis() {
	redisUrl := os.Getenv("REDIS_HOST")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})
	ctx = context.Background()
}

// AddToQueue adds an item to a Redis sorted set (queue) with a given score
func AddToRedisQueue(queueName string, member string, score float64) error {
	z := &redis.Z{
		Score:  score,
		Member: member,
	}
	if err := redisClient.ZAdd(ctx, queueName, *z).Err(); err != nil {
		return fmt.Errorf("failed to ZADD to queue %s: %w", queueName, err)
	}
	return nil
}
