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

	fmt.Println("--- TABLES ---")
	tRows, _ := conn.Query(context.Background(), "SELECT id, number, name, status FROM dining_tables ORDER BY id")
	for tRows.Next() {
		var id int64
		var num int32
		var name, status string
		tRows.Scan(&id, &num, &name, &status)
		fmt.Printf("ID: %d | Num: %d | Name: %s | Status: %s\n", id, num, name, status)
	}

	fmt.Println("\n--- MENU ---")
	mRows, _ := conn.Query(context.Background(), "SELECT id, name FROM menu_items ORDER BY id")
	for mRows.Next() {
		var id int64
		var name string
		mRows.Scan(&id, &name)
		fmt.Printf("ID: %d | Name: %s\n", id, name)
	}

	fmt.Println("\n--- ORDERS ---")
	oRows, _ := conn.Query(context.Background(), "SELECT id, table_id, status FROM orders ORDER BY id DESC LIMIT 20")
	for oRows.Next() {
		var id, tid int64
		var status string
		oRows.Scan(&id, &tid, &status)
		fmt.Printf("ID: %d | TableID: %d | Status: %s\n", id, tid, status)
	}
}
