package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// hardcode data source name rather than reading from environment to simplify demo
// Postgres default port is 5432, so pick localhost port 5433 to avoid colliding with a local Postgres on 5432
const dsn = "postgres://postgres:demo@localhost:5433/paydemo"

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// 1. create a table and insert a new payment row
	// clean the table for last run
	// a field = name + type + optional constraints
	// idem_key is identifier used to locate one payment
	// status holds the state machine state
	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS payments;
		CREATE TABLE payments (
			id		SERIAL PRIMARY KEY,
			idem_key TEXT UNIQUE NOT NULL,
			status  TEXT NOT NULL,
			amount	INT NOT NULL
		);
		INSERT INTO payments (idem_key, status, amount)
    	VALUES ('key_123', 'created', 100);
	`)
	if err != nil {
		panic(err)
	}

	// 2. [legal] created -> processing
	tag, err := pool.Exec(ctx, `
		UPDATE payments
    	SET status='processing'
		WHERE idem_key='key_123' AND status='created';
	`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[legal] created -> processing, RowsAffected: %d\n", tag.RowsAffected())

	// 3. [legal] processing -> succeeded (terminal state)
	tag, err = pool.Exec(ctx, `
		UPDATE payments
    	SET status='succeeded'
		WHERE idem_key='key_123' AND status='processing';
	`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[legal] processing -> succeeded, RowsAffected: %d\n", tag.RowsAffected())

	// 4. [illegal] a duplicate success callback
	tag, err = pool.Exec(ctx, `
		UPDATE payments
    	SET status='succeeded'
		WHERE idem_key='key_123' AND status='processing';
	`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[illegal] a duplicate callback, RowsAffected: %d\n", tag.RowsAffected())

	// 5. check the final state, the terminal state is not corrupted
	var finalStatus string
	err = pool.QueryRow(ctx,
		`SELECT status FROM payments WHERE idem_key='key_123'`,
	).Scan(&finalStatus)
	if err != nil {
		panic(err)
	}
	fmt.Printf("final status: %s\n", finalStatus)
}
