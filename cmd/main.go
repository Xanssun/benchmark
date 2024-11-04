package main

import (
	"log"

	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Xanssun/benchmark.git/internal/db"
)

type config struct {
	db             dbConfig
	durationMillis int
}

type dbConfig struct {
	dsn   string
	query string
}

func benchRPS(database *sql.DB, query string, duration time.Duration) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 0
	endTime := time.Now().Add(duration)

	executeQuery := func() {
		defer wg.Done()
		for time.Now().Before(endTime) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := database.QueryContext(ctx, query)
			if err == nil {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}
	}

	// Колличество горутин
	numGoroutines := 100
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go executeQuery()
	}
	wg.Wait()

	return count * 1000 / int(duration.Milliseconds())
}

func main() {
	cfg := config{
		durationMillis: 1000, // Время
		db: dbConfig{
			dsn:   "postgres://postgres:postgres@localhost:55432/northwind?sslmode=disable",
			query: "SELECT o.order_id, o.customer_id, o.order_date, od.product_id, od.unit_price, od.quantity, od.discount FROM orders o JOIN order_details od ON o.order_id = od.order_id;", // Пример запроса
		},
	}

	database, err := db.New(cfg.db.dsn)
	if err != nil {
		log.Panic(err)
	}
	defer database.Close()

	log.Printf("Установлено подключение к базе данных")

	rps := benchRPS(database, cfg.db.query, time.Duration(cfg.durationMillis)*time.Millisecond)
	log.Printf("RPS: %d\n", rps)
}
