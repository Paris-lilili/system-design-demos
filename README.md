## Verify result
```
(launch docker desktop firstly)

❯ docker run --name pay-demo-pg -e POSTGRES_PASSWORD=demo -e POSTGRES_DB=paydemo -p 5433:5432 -d postgres:16

❯ go run main.go
inserts times:      1
the number of rows in table for idem_key: 1
Success: only one charge was recorded
```
