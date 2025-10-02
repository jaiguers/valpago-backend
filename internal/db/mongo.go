package db

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var (
    mongoClient *mongo.Client
    mongoDB     *mongo.Database
)

func ConnectMongo(uri, db string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil { return err }
    if err := client.Ping(ctx, nil); err != nil { return err }
    mongoClient = client
    mongoDB = client.Database(db)
    return nil
}

func Mongo() *mongo.Database { return mongoDB }

