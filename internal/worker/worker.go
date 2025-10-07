package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/usuario/valpago-backend/internal/config"
	"github.com/usuario/valpago-backend/internal/db"
)

func Start() {
	if db.Rdb == nil {
		log.Println("Redis not available, skipping worker...")
		return
	}

	log.Println("Starting Redis worker...")

	ctx := context.Background()
	groupName := config.C.RedisGroup
	streamName := config.C.RedisStreamNS
	consumerName := fmt.Sprintf("%s-%d", config.C.RedisConsumer, time.Now().Unix())

	// Create consumer group if it doesn't exist
	err := db.Rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Printf("Error creating consumer group %s for stream %s: %v", groupName, streamName, err)
	}

	for {
		// Read from Redis stream
		streams, err := db.Rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    time.Second * 5,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// No messages, continue
				continue
			}
			log.Printf("Redis stream error: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				// Process transaction
				processTransaction(ctx, message)

				// Acknowledge message
				db.Rdb.XAck(ctx, streamName, groupName, message.ID)
			}
		}
	}
}

func processTransaction(ctx context.Context, message redis.XMessage) {
	log.Printf("Processing transaction: %s", message.ID)

	// Simulate processing time
	time.Sleep(time.Second * 2)

	// Notificar al front solo los PENDING (para que aparezcan en la cola)
	if raw, ok := message.Values["data"]; ok {
		payload := fmt.Sprintf("%v", raw)
		notification := map[string]interface{}{
			"type":      "transaction.pending",
			"data":      payload,
			"timestamp": time.Now().Unix(),
		}
		db.Rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: config.C.RedisNotificationsStream,
			Values: notification,
		})
		return
	}

	// Fallback: publicar evento mínimo si no hay "data"
	notificationData := map[string]interface{}{
		"status":    "REVIEW",
		"timestamp": time.Now().Unix(),
	}
	db.Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: config.C.RedisNotificationsStream,
		Values: notificationData,
	})
}
