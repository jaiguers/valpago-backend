package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func ConnectRedis(url string) error {
	fmt.Printf("Connecting to Redis: %s\n", url)

	// Parse the Redis URL directly
	opt, err := redis.ParseURL(url)
	if err != nil {
		fmt.Printf("Error parsing Redis URL: %v\n", err)
		return err
	}

	Rdb = redis.NewClient(opt)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test the connection
	if err := Rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	fmt.Println("Redis connected successfully")
	return nil
}
