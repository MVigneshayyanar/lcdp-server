package scratch

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CheckCounts() {
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var count int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM inventory_items").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inventory Count: %d\n", count)
	
	var billCount int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM bills").Scan(&billCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bill Count: %d\n", billCount)
}
