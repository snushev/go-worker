package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbConfig struct{ DB *pgxpool.Pool }

func main() {
	dsn := "postgres://postgres:password@localhost:5432/go-worker"

	pool := NewPostgresPool(dsn)

	apiCfg := dbConfig{DB: pool}

	var wg sync.WaitGroup

	mux := http.NewServeMux()

	mux.HandleFunc("POST /job", apiCfg.HandlerJobCreate)
	mux.HandleFunc("GET /jobs", apiCfg.HandlerJobsList)

	go func() {
		fmt.Println("Server is up on port :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker(i, pool)
		}(i)
	}
	wg.Wait()

}

func worker(workerId int, cfg *pgxpool.Pool) {
	ctx := context.Background()
	fmt.Printf("Worker %d starting\n", workerId)
	for {
		id, jsonData, err := fetchAndLockJob(ctx, cfg)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				fmt.Println("No jobs available, sleeping...")
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Printf("Fetch error: %v\n", err)
			continue
		}

		fmt.Printf("[Worker %d] JOB LOCKED: %s\n", workerId, jsonData)

		var payload JobPayload
		err = json.Unmarshal(jsonData, &payload)
		if err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			markFailed(ctx, cfg, id)
			continue
		}

		executeJob(workerId, payload)

		err = markDone(ctx, cfg, id)
		if err != nil {
			fmt.Printf("Mark done error: %v\n", err)
		}
	}
}

func fetchAndLockJob(ctx context.Context, db *pgxpool.Pool) (int, []byte, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, nil, err
	}

	var (
		id   int
		data []byte
	)

	query := `
		SELECT id, payload 
		FROM jobs 
		WHERE status = 'pending'
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	err = tx.QueryRow(ctx, query).Scan(&id, &data)
	if err != nil {
		tx.Rollback(ctx)
		return 0, nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE jobs SET status = 'processing' WHERE id = $1
	`, id)
	if err != nil {
		tx.Rollback(ctx)
		return 0, nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, nil, err
	}

	return id, data, nil
}

func executeJob(id int, p JobPayload) {
	switch p.Task {
	case "print":
		fmt.Printf("[Worker %d] PRINT: %s\n", id, p.Value)

	case "sleep":
		fmt.Printf("[Worker %d] SLEEP: %d seconds\n", id, p.Seconds)
		time.Sleep(time.Duration(p.Seconds) * time.Second)

	default:
		fmt.Println("Unknown task")
	}
}

func markDone(ctx context.Context, db *pgxpool.Pool, id int) error {
	_, err := db.Exec(ctx,
		`UPDATE jobs SET status = 'done' WHERE id = $1`,
		id,
	)
	return err
}

func markFailed(ctx context.Context, db *pgxpool.Pool, id int) {
	_, _ = db.Exec(ctx,
		`UPDATE jobs SET status = 'failed' WHERE id = $1`,
		id,
	)
}
