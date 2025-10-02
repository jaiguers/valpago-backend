package config

import (
    "errors"
    "os"
    "strconv"
)

type Config struct {
    MongoURI       string
    MongoDB        string
    RedisURL       string
    RedisStreamNS  string
    RedisGroup     string
    RedisConsumer  string
    JWTSecret      string
    JWTExpHours    int
    APIKeyHeader   string
    ServerPort     int
    AllowedOrigins string
}

var C Config

func Load() error {
    C.MongoURI = os.Getenv("MONGODB_URI")
    if C.MongoURI == "" { return errors.New("MONGODB_URI is required") }
    C.MongoDB = getenv("MONGODB_DB", "valpago")
    C.RedisURL = getenv("REDIS_URL", "redis://127.0.0.1:6379")
    C.RedisStreamNS = getenv("REDIS_STREAM_NAMESPACE", "valpago:transactions")
    C.RedisGroup = getenv("REDIS_CONSUMER_GROUP", "valpago:cg")
    C.RedisConsumer = getenv("REDIS_CONSUMER_NAME", "worker-1")
    C.JWTSecret = getenv("JWT_SECRET", "dev_secret_change_me")
    C.JWTExpHours = getenvInt("JWT_EXP_HOURS", 24)
    C.APIKeyHeader = getenv("API_KEY_HEADER_NAME", "x-api-key")
    C.ServerPort = getenvInt("SERVER_PORT", 8080)
    C.AllowedOrigins = getenv("ALLOWED_ORIGINS", "*")
    return nil
}

func getenv(key, def string) string {
    v := os.Getenv(key)
    if v == "" { return def }
    return v
}

func getenvInt(key string, def int) int {
    v := os.Getenv(key)
    if v == "" { return def }
    n, err := strconv.Atoi(v)
    if err != nil { return def }
    return n
}

