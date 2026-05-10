package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	url := os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	conn.Exec(context.Background(), "DELETE FROM orders")
	conn.Exec(context.Background(), "UPDATE dining_tables SET status = 'available'")

	fmt.Println("Database reset successful")
}
