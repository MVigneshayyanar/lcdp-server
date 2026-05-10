package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/Maestrominds/lcdp-restaurant/internal/api"
	"github.com/Maestrominds/lcdp-restaurant/internal/config"
	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		fmt.Printf("[NETWORK] %s %s from %s\n", c.Method(), c.Path(), c.IP())
		return c.Next()
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))

	handler := api.NewHandler(cfg, db.New(pool), pool)
	handler.Register(app)

	log.Fatal(app.Listen(cfg.Addr))
}
