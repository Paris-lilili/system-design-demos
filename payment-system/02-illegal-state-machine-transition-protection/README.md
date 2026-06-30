## Payment lifecycle
The payment moves through state machine. Only certain transitions are legal:
```
created ──► processing ──► succeeded
                      └──► failed
```
Note: `succeeded` and  `failed` are termial, once arrived, the state cannot change again

## Real flow
1. Customer clicks pay, the application create a payment row with status `created`.
2. Application call PSP(payment service provider) to handle the pay
3. PSP accepts the request and start processing, the application move the local status to `processing`
4. PSP finish and send a callback (aka webhook) back to the application about the result success or fail
5. Once receiving the callback, the application update the local row to `succeeded` or `failed`

A production system protects state transitions in **two layers**:
 
- **Application layer** — judges the *business logic*: is this a duplicate callback? an out-of-order callback? what error or retry should we return?
- **Database layer** — guarantees *concurrency correctness*: the legal predecessor state is encoded into the `UPDATE` itself, so the final write is atomic.

**This demo omits the application layer.** It simplifies the app to: connect to the database, run SQL, check the result. 
The goal is to isolate and prove the database guard on its own state transition correctness here relies entirely on the atomic `UPDATE`

## What this verifies
The illegal state transition cannot corrupt the record. e.g. step 5 callback might arrive late, out of order, resend more than once etc..

1. Insert a payment row.
2. Legal transition: run a transition whose WHERE matches the current state. Expect RowsAffected = 1, state advances.
3. Illegal transition: force the row to a terminal state (succeeded), then run a transition that requires an earlier predecessor 
(e.g. back to processing). Expect RowsAffected = 0, state no change.

## Output
```
❯ go run main.go
[legal] created -> processing, RowsAffected: 1
[legal] processing -> succeeded, RowsAffected: 1
[illegal] a duplicate callback, RowsAffected: 0
final status: succeeded
```