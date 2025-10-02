package sse

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/usuario/valpago-backend/internal/config"
	"github.com/usuario/valpago-backend/internal/db"
)

func Register(e *echo.Echo) {
	e.GET("/api/sse", handleSSE)
}

func handleSSE(c echo.Context) error {
	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create a channel to send messages
	messageChan := make(chan string, 10)
	defer close(messageChan)

	// Start Redis stream consumer in a goroutine
	go func() {
		ctx := context.Background()
		
		// Create consumer group if it doesn't exist
		groupName := config.C.RedisGroup
		streamName := config.C.RedisStreamNS
		consumerName := fmt.Sprintf("%s-%d", config.C.RedisConsumer, time.Now().Unix())

		// Try to create consumer group (ignore error if it already exists)
		db.Rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0")

		for {
		// Read from Redis stream
		streams, err := db.Rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    1,
			Block:    time.Second * 5,
		}).Result()

			if err != nil {
				if err == redis.Nil {
					// No messages, continue
					continue
				}
				fmt.Printf("Redis stream error: %v\n", err)
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					// Format message for SSE
					data := fmt.Sprintf("data: %s\n\n", message.Values["status"])
					select {
					case messageChan <- data:
					default:
						// Channel full, skip message
					}
				}
			}
		}
	}()

	// Send messages to client
	for {
		select {
		case message := <-messageChan:
			if _, err := c.Response().Write([]byte(message)); err != nil {
				return err
			}
			c.Response().Flush()
		case <-c.Request().Context().Done():
			return nil
		}
	}
}
