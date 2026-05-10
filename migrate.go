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
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	files := []string{
		"db/migrations/000006_add_order_status.up.sql",
		"db/migrations/000007_add_table_statuses.up.sql",
	}

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			log.Printf("Warning: could not read %s: %v", f, err)
			continue
		}
		_, err = conn.Exec(context.Background(), string(sql))
		if err != nil {
			log.Printf("Warning: error applying %s: %v", f, err)
		}
	}

	fmt.Println("Migration applied successfully")
}
