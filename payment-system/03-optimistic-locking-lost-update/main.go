package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// hardcode data source name rather than reading from environment to simplify demo
// Postgres default port is 5432, so pick localhost port 5433 to avoid colliding with a local Postgres on 5432
const dsn = "postgres://postgres:demo@localhost:5433/paydemo"

func main() {
	ctx := context.Background()

	// Exec() only execute sql without return data - create table, update value
	// QueryRow()+Scan(&a, &b) execute sql and read the value to go varibles
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
	// version is for optimistic locking to identify if any other modification ahead of this
	// amount originally is 100, the goal is deduct 20, terminal amount should be 80
	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS payments;
		CREATE TABLE payments (
			id		SERIAL PRIMARY KEY,
			idem_key TEXT UNIQUE NOT NULL,
			status  TEXT NOT NULL,
			amount	INT NOT NULL,
			version INT NOT NULL
		);
		INSERT INTO payments (idem_key, status, amount, version)
    	VALUES ('key_123', 'created', 100, 1);
	`)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// 2. one goroutine simulate retry logic: after timeout, retry failed. 
	go func() {
		defer wg.Done()

		// 2.1 read version and status
		var version int
		var status string

		err := pool.QueryRow(ctx, `
    		SELECT version, status
			FROM payments
			WHERE idem_key='key_123';
		`).Scan(&version, &status)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Goroutine1 read, version= %d, status= %q\n", version, status)

		// 2.2 sleep 5 second to simulate after reading the status is processing, PSP is handling, but app level retry
		// retry and recall two different pathes both try to update value(this routine is retry path)
		time.Sleep(3 * time.Second)

		// 2.3 simulate retry - expect **failed** and RowsAffected suppose to be 0
		tag, err := pool.Exec(ctx, `
			UPDATE payments
			SET status='succeeded', amount=amount-20, version=version+1
			WHERE idem_key='key_123' AND version=$1;
		`, version)
		if err != nil {
			panic(err)
		}
		fmt.Printf("After retry, RowsAffected: %d\n", tag.RowsAffected())
	}()

	// 3. another goroutine PSP **successfully** deduct money(callback path) ahead of retry finished
	go func() {
		defer wg.Done()

		// sleep 0.5s to make sure goroutine2 start later than goroutine1
		time.Sleep(500*time.Millisecond)

		var version int
		var status string

		err := pool.QueryRow(ctx, `
    		SELECT version, status
			FROM payments
			WHERE idem_key='key_123';
		`).Scan(&version, &status)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Goroutine2 read, version= %d, status= %q\n", version, status)

		tag, err := pool.Exec(ctx, `
			UPDATE payments
			SET status='succeeded', amount=amount-20, version=version+1
			WHERE idem_key='key_123' AND version=$1;
		`, version)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Recall path, RowsAffected: %d\n", tag.RowsAffected())
	}()

	wg.Wait()

	// 4. verify version is 2, amount is 80(only deduct money once)
	var version int
	var amount int
	err = pool.QueryRow(ctx, 
		`SELECT version, amount FROM payments WHERE idem_key='key_123'`,
	).Scan(&version, &amount)
	if err != nil {
		panic(err)
	}

	fmt.Printf("expect version is 2, actual version is: %d\n", version)
	fmt.Printf("expect amount is 80, actual amount is: %d\n", amount)
}
