package main

import (
	"context"
	"fmt"
	"log"

	"strings"
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
	godotenv.Load()

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
		err := c.Next()
		if setCookie := c.Response().Header.Peek("Set-Cookie"); len(setCookie) > 0 {
			fmt.Printf("[DEBUG] Set-Cookie: %s\n", string(setCookie))
		}
		return err
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// In development, be permissive to allow Flutter Web (random ports), Svelte (5173), and Android Emulator
			return cfg.Env == "local" || strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "http://10.0.2.2") || origin == ""
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Cookie",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	handler := api.NewHandler(cfg, db.New(pool), pool)
	handler.Register(app)

	log.Fatal(app.Listen(cfg.Addr))
}
