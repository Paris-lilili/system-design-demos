## What this verifies
A payment system ledger records where money moves and it have to always stay balanced. The rule keeps double-entry bookkeeping: every business event produces at least two entries, one debit and one credit, equal in amount, opposite in direction.

1. sum debit equals to sum credit after several payment and refunds
2. A refund never exceeds the refundable amount
3. A rejected over-refund does not record to the ledger

## Scenario
pay    100   (customer -> shop)
refund  30   (shop -> customer)
refund  60   (shop -> customer)   # cumulative 90 <= 100, ok
refund  80   (shop -> customer)   # only 10 refundable -> rejected

## How to Run
```
❯ go run main.go
Success: debit == credit
Success: over-refund rejected, ledger untouched
```
