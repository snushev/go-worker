package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbConfig struct{ DB *pgxpool.Pool }

func main() {
	dsn := "postgres://postgres:password@localhost:5432/go-worker"

	pool := NewPostgresPool(dsn)
	cfg := dbConfig{DB: pool}

	ctx := context.Background()

	for {
		tx, err := cfg.DB.Begin(ctx)
		if err != nil {
			fmt.Printf("Transaction begin error: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var (
			id       int
			jsonData []byte
		)

		query := `
			SELECT id, payload 
			FROM jobs 
			WHERE status = 'pending' 
			FOR UPDATE SKIP LOCKED 
			LIMIT 1
		`

		err = tx.QueryRow(ctx, query).Scan(&id, &jsonData)
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("No jobs available, sleeping...")
			tx.Rollback(ctx)
			time.Sleep(2 * time.Second)
			continue
		}
		if err != nil {
			fmt.Printf("Query error: %v\n", err)
			tx.Rollback(ctx)
			continue
		}

		updateQuery := `
			UPDATE jobs 
			SET status = 'processing' 
			WHERE id = $1
		`

		_, err = tx.Exec(ctx, updateQuery, id)
		if err != nil {
			fmt.Printf("Update error: %v\n", err)
			tx.Rollback(ctx)
			continue
		}

		err = tx.Commit(ctx)
		if err != nil {
			fmt.Printf("Commit error: %v\n", err)
			tx.Rollback(ctx)
			continue
		}

		fmt.Printf("JOB LOCKED & PROCESSING: %s\n", jsonData)

	}
}
