package worker

import (
	"context"
	"fmt"
	"log"
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
	if err != nil {
		log.Printf("Error creating consumer group: %v", err)
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

	// Update transaction status to REVIEW
	transactionID := message.Values["transaction_id"]
	if transactionID != nil {
		// In a real implementation, you would update the MongoDB transaction
		// and send notifications via SSE
		log.Printf("Transaction %s processed, status updated to REVIEW", transactionID)

		// Send notification to SSE stream
		notificationData := map[string]interface{}{
			"transaction_id": transactionID,
			"status":         "REVIEW",
			"timestamp":      time.Now().Unix(),
		}
		db.Rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "valpago:notifications",
			Values: notificationData,
		})
	}
}
