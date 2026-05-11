package scratch

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func DebugRevenue() {
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("Running GetWeeklyRevenue test...")
	rows, err := pool.Query(context.Background(), `
		WITH days AS (
			SELECT generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, '1 day')::date AS d
		)
		SELECT 
			TO_CHAR(d, 'Dy') as day_name,
			COALESCE(SUM(m.price * o.quantity), 0)::FLOAT as revenue,
			d as day_date
		FROM days d
		LEFT JOIN orders o ON o.ordered_at::date = d
		LEFT JOIN menu_items m ON o.menu_item_id = m.id
		GROUP BY d
		ORDER BY d;
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var rev float64
		var date any
		rows.Scan(&name, &rev, &date)
		fmt.Printf("Day: %s, Revenue: %.2f, Date: %v\n", name, rev, date)
	}
}
