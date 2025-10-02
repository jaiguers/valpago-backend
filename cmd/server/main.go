package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/usuario/valpago-backend/internal/config"
	"github.com/usuario/valpago-backend/internal/db"
	"github.com/usuario/valpago-backend/internal/routes"
	"github.com/usuario/valpago-backend/internal/sse"
	"github.com/usuario/valpago-backend/internal/worker"
)

func main() {
	_ = godotenv.Load()

	if err := config.Load(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	log.Printf("Connecting to MongoDB: %s", config.C.MongoURI)
	if err := db.ConnectMongo(config.C.MongoURI, config.C.MongoDB); err != nil {
		log.Fatalf("mongo error: %v", err)
	}
	log.Println("MongoDB connected successfully")

	log.Printf("Connecting to Redis: %s", config.C.RedisURL)
	if err := db.ConnectRedis(config.C.RedisURL); err != nil {
		log.Printf("redis warning: %v (continuing without Redis)", err)
	} else {
		log.Println("Redis connected successfully")
	}

	go worker.Start()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{config.C.AllowedOrigins},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization, config.C.APIKeyHeader},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions},
	}))

	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
	routes.Register(e)
	sse.Register(e)

	port := config.C.ServerPort
	if p := os.Getenv("SERVER_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("listening on %s", addr)
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
