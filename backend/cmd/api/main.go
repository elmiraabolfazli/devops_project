// backend/cmd/api/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	app *fiber.App
	rdb *redis.Client
	ctx context.Context
}

func NewServer() *Server {
	// بارگذاری متغیرهای محیطی
	_ = godotenv.Load()

	// تنظیمات Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()

	// تست اتصال به Redis
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Warning: Redis connection failed: %v", err)
		log.Println("📝 The app will continue but without Redis cache")
	} else {
		log.Println("✅ Redis connected successfully")
	}

	// تنظیمات Fiber
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.Next()
	})

	s := &Server{
		app: app,
		rdb: rdb,
		ctx: ctx,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Health check
	s.app.Get("/health", s.healthCheck)

	// API routes
	api := s.app.Group("/api")
	api.Post("/data", s.createData)
	api.Get("/data", s.getAllData)
	api.Get("/data/:id", s.getData)
	api.Delete("/data/:id", s.deleteData)
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Service is healthy",
		"data": fiber.Map{
			"status": "running",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

func (s *Server) createData(c *fiber.Ctx) error {
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	id := fmt.Sprintf("data:%d", time.Now().UnixNano())
	data, err := json.Marshal(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to marshal data",
		})
	}

	if err := s.rdb.Set(s.ctx, id, data, 0).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save data: " + err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Data created successfully",
		"data": fiber.Map{
			"id": id,
		},
	})
}

func (s *Server) getAllData(c *fiber.Ctx) error {
	keys, err := s.rdb.Keys(s.ctx, "data:*").Result()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve keys: " + err.Error(),
		})
	}

	results := make(map[string]interface{})
	for _, key := range keys {
		data, err := s.rdb.Get(s.ctx, key).Result()
		if err == nil {
			var value interface{}
			json.Unmarshal([]byte(data), &value)
			results[key] = value
		}
	}

	return c.JSON(fiber.Map{
		"message": "All data retrieved successfully",
		"data":    results,
	})
}

func (s *Server) getData(c *fiber.Ctx) error {
	id := c.Params("id")

	data, err := s.rdb.Get(s.ctx, id).Result()
	if err == redis.Nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Data not found",
		})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve data: " + err.Error(),
		})
	}

	var result interface{}
	json.Unmarshal([]byte(data), &result)

	return c.JSON(fiber.Map{
		"message": "Data retrieved successfully",
		"data":    result,
	})
}

func (s *Server) deleteData(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.rdb.Del(s.ctx, id).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete data: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Data deleted successfully",
	})
}

func (s *Server) Run(port string) error {
	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", port)
		log.Printf("📋 Available endpoints:")
		log.Printf("   GET    /health")
		log.Printf("   GET    /api/data")
		log.Printf("   POST   /api/data")
		log.Printf("   GET    /api/data/:id")
		log.Printf("   DELETE /api/data/:id")

		if err := s.app.Listen(":" + port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("🛑 Shutting down server...")

	// Close Redis connection
	if err := s.rdb.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	// Shutdown Fiber app
	return s.app.Shutdown()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewServer()
	if err := server.Run(port); err != nil {
		log.Fatal(err)
	}
}
