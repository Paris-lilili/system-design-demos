## What this verifies

When the same payment request arrives concurrently (a customer double-clicks,
or a client retries after a timeout), the system must record the charge exactly once. 
Otherwise the customer is charged duplicated.

Note: 
- `idem_key` dedups payment creation (UNIQUE constraint, at INSERT).
- `version` prevents lost update (optimistic locking, at UPDATE).
They guard different failure modes: duplicate execution vs concurrent overwrite.

## How to Run
```
(launch docker desktop firstly)

❯ docker run --name pay-demo-pg -e POSTGRES_PASSWORD=demo -e POSTGRES_DB=paydemo -p 5433:5432 -d postgres:16

❯ go run main.go
inserts times:      1
the number of rows in table for idem_key: 1
Success: only one charge was recorded
```
