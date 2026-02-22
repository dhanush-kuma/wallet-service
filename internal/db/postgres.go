package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	var pool *pgxpool.Pool
	var err error

	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(context.Background(), dsn)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				fmt.Println("Connected to DB")
				return pool, nil
			}
		}

		fmt.Println("waiting for database...")
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to db after retries: %w", err)
}