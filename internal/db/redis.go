package db

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func ConnectRedis(url string) error {
	// Parse the Redis URL directly
	opt, err := redis.ParseURL(url)
	if err != nil {
		return err
	}

	// If it's a TLS connection (rediss://), configure TLS
	if strings.HasPrefix(url, "rediss://") {
		opt.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}
	}

	Rdb = redis.NewClient(opt)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return Rdb.Ping(ctx).Err()
}
