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

	// Use a cancellable context tied to the request
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	// Start Redis stream consumer (broadcast) in a goroutine
	go func() {

		streamName := config.C.RedisNotificationsStream
		lastID := "$" // empezar desde nuevos mensajes

		for {
			// Stop if client disconnected
			select {
			case <-ctx.Done():
				return
			default:
			}
			// Read from Redis stream (XREAD broadcast)
			streams, err := db.Rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{streamName, lastID},
				Count:   10,
				Block:   time.Second * 5,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// No messages, continue
					continue
				}
				// If context was cancelled, exit
				if ctx.Err() != nil {
					return
				}
				fmt.Printf("Redis stream error: %v\n", err)
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					// Preferir el JSON completo si viene en "data"; si no, construir uno simple
					var payload string
					if raw, ok := message.Values["data"]; ok {
						payload = fmt.Sprintf("%v", raw)
					} else {
						payload = fmt.Sprintf(`{"status":"%v"}`, message.Values["status"])
					}
					data := fmt.Sprintf("data: %s\n\n", payload)
					select {
					case messageChan <- data:
					default:
						// Channel full, skip message
					}
					lastID = message.ID
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
			// Ensure goroutine stops
			cancel()
			return nil
		}
	}
}
