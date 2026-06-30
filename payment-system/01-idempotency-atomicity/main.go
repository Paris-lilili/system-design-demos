package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
)

// hardcode data source name rather than reading from environment to simplify demo
// Postgres default port is 5432, so localhost port random pick a close number
const dsn = "postgres://postgres:demo@localhost:5433/paydemo"

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// 1. create a table
	// clean the table for last run
	// a field = name + type + optional constraints
	// idem_key with UNIQUE constraint because it should reject duplication
	// id type is SERIAL that can auto-increment integer
	// Exec executes the given SQL without return data
	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS payments;
		CREATE TABLE payments (
			id		SERIAL PRIMARY KEY,
			idem_key TEXT UNIQUE NOT NULL,
			status  TEXT NOT NULL,
			amount	INT NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	}

	// 2. mock 200 concurrent requests, which carry the same idem_key. (too many click or retry)
	var wg sync.WaitGroup
	var successInsert int64

	for i := 0; i < 200; i++ {
		wg.Add(1)

		// 3. all goroutine execute insert sql, using successInsert as counter to check how many successful insert
		go func() {
			defer wg.Done()

			// QueryRow executes a query that is expected to return at most one row(return data).
			// Scan read the return data into insertedID, if query failed -> no rows were found, it returns ErrNoRows.
			// simulate duplicate insert, so insert value is fixed
			var insertedID int
			err := pool.QueryRow(ctx, `
				INSERT INTO payments (idem_key, status, amount)
				VALUES ($1, $2, $3)
				ON CONFLICT (idem_key) DO NOTHING
				RETURNING id;
			`, "key_123", "created", 100).Scan(&insertedID)

			if err == nil {
				// scan successfully get return data, which means this gorotine successfully insert one row
				atomic.AddInt64(&successInsert, 1)
			}
		}()
	}

	wg.Wait()

	// 4. verify only one insert success and only one row for idem_key("key_123")
	var rowCount int
	err = pool.QueryRow(ctx,
		`SELECT count(*) FROM payments WHERE idem_key = 'key_123'`,
	).Scan(&rowCount)
	if err != nil {
		panic(err)
	}

	fmt.Printf("inserts times: %d\n", successInsert)
	fmt.Printf("the number of rows in table for idem_key: %d\n", rowCount)

	if successInsert == 1 && rowCount == 1 {
		fmt.Println("Success: only one payment was recorded")
	} else {
		fmt.Println("Fail: idempotency broke, duplicate or zero charges")
	}
}
