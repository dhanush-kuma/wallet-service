# Wallet Service (Go + PostgreSQL)


A simple high‑concurrency wallet backend built in **Go**, demonstrating:


-   Double‑entry ledger
-   Idempotent transactions
-   Deadlock‑safe transfers
-   PostgreSQL with pgx
-   Dockerized deployment
-   Deployed on Railway


## Tech choice
 Language: Go (Golang) because of its ability to handle high-concurrency, low-latency transaction processing safely. Also I wanted to learn Go.


 RDBMS: PostgreSQL because of its Strict ACID properties and features like Row Locking to prevent race conditions, and ability to detect deadlocks.


------------------------------------------------------------------------


## Live Preview


Live base URL:


    https://wallet-service-production-3d33.up.railway.app
sample get balance url: https://wallet-service-production-3d33.up.railway.app/wallets/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb/balance


------------------------------------------------------------------------


## Database Schema


Main tables:


-   users
-   assets
-   wallets
-   transactions
-   ledger_entries


------------------------------------------------------------------------


## Local Development


### 1. Clone repo


    git clone <repo-url>
    cd wallet-service


### 2. Run with Docker


    docker compose up --build


This will:


-   start PostgreSQL
-   run migrations
-   run seed
-   API hardcoded on port 8080


------------------------------------------------------------------------


## Environment Variables


Required:


    DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require


------------------------------------------------------------------------


## API Endpoints for Testing with Postman


### Get balance


    GET /wallets/:wallet_id/balance


------------------------------------------------------------------------


### Top up wallet


    POST /wallets/:wallet_id/topup


Body:


``` json
{
  "reference_id": "txn-001",
  "amount": 1000,
  "asset_code": "gold"
}
```


------------------------------------------------------------------------
### Reward a wallet


    POST /wallets/:wallet_id/bonus


Body:


``` json
{
  "reference_id": "rwd-001",
  "amount": 100,
  "asset_code": "gold"
}
```


------------------------------------------------------------------------
### Redemption of wallet asset


    POST /wallets/:wallet_id/spend


Body:


``` json
{
  "reference_id": "wdr-001",
  "amount": 150,
  "asset_code": "gold"
}
```


------------------------------------------------------------------------


### Create user


    POST /users


  Body:


``` json
{
  "name": "Armin"
}
```
------------------------------------------------------------------------


### Create wallet


    POST /wallets


  Body:


``` json
{
  "label": "Armin Gold wallet",
  "user_id": "e1e1e1e1-e1e1-e1e1-e1e2-e1e1e1e1e1e2",
  "asset_type_id": 1
}
```


------------------------------------------------------------------------


### Create asset


    POST /assets


  Body:


``` json
{
  "code": "SILVER"
}
```

------------------------------------------------------------------------


### Get transactions 


    GET /transactions?limit=20&offset=0


------------------------------------------------------------------------


### Get Ledger_entries


    GET /ledger-entries?limit=20&offset=0

------------------------------------------------------------------------


## Idempotency


All money‑moving endpoints require:


    reference_id


If the same reference is sent twice:


second call returns previous result\
no incorrect double entry


------------------------------------------------------------------------



# Core of all transactions in system:


``` go
func (r *Repository) Transfer(
    ctx context.Context,
    referenceID string,
    fromWalletID uuid.UUID,
    toWalletID uuid.UUID,
    amount int64,
) error {


    if amount <= 0 {
        return fmt.Errorf("amount must be positive")
    }


    tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)




    // check if transaction already completed


    var existingID uuid.UUID
    err = tx.QueryRow(ctx,
        `SELECT id FROM transactions WHERE reference_id = $1`,
        referenceID,
    ).Scan(&existingID)


    if err == nil {
        return tx.Commit(ctx)
    }
    if !errors.Is(err, pgx.ErrNoRows) {
        return err
    }




    // Lock wallets in deterministic order


    var first, second uuid.UUID
    if fromWalletID.String() < toWalletID.String() {
        first, second = fromWalletID, toWalletID
    } else {
        first, second = toWalletID, fromWalletID
    }


    // lock first wallet
    if _, err := tx.Exec(ctx,
        `SELECT id FROM wallets WHERE id = $1 FOR UPDATE`,
        first,
    ); err != nil {
        return err
    }


    // lock second wallet
    if _, err := tx.Exec(ctx,
        `SELECT id FROM wallets WHERE id = $1 FOR UPDATE`,
        second,
    ); err != nil {
        return err
    }




    // Check balance


    var balance int64
    err = tx.QueryRow(ctx,
        `SELECT balance FROM wallets WHERE id = $1`,
        fromWalletID,
    ).Scan(&balance)
    if err != nil {
        return err
    }


    if balance < amount {
        return fmt.Errorf("insufficient balance")
    }




    // Create transaction record


    txnID := uuid.New()


    _, err = tx.Exec(ctx,
        `INSERT INTO transactions (id, reference_id, type, status)
         VALUES ($1, $2, 'transfer', 'completed')`,
        txnID,
        referenceID,
    )
    if err != nil {
        return err
    }




    // Insert ledger entries (double entry)


    debitEntryID := uuid.New()
    creditEntryID := uuid.New()


    _, err = tx.Exec(ctx,
        `INSERT INTO ledger_entries
            (id, transaction_id, wallet_id, direction, amount)
         VALUES
            ($1, $2, $3, 'debit', $4),
            ($5, $2, $6, 'credit', $4)`,
        debitEntryID,
        txnID,
        fromWalletID,
        amount,
        creditEntryID,
        toWalletID,
    )
    if err != nil {
        return err
    }




    // Update cached balances


    _, err = tx.Exec(ctx,
        `UPDATE wallets
         SET balance = balance - $1
         WHERE id = $2`,
        amount,
        fromWalletID,
    )
    if err != nil {
        return err
    }


    _, err = tx.Exec(ctx,
        `UPDATE wallets
         SET balance = balance + $1
         WHERE id = $2`,
        amount,
        toWalletID,
    )
    if err != nil {
        return err
    }




    return tx.Commit(ctx)
}
```
function transfer is used in all transaction related function, this function moves assets from one wallet to another, it checks if amount to be transfered is greater than 0, begins a db transaction, 

## Idempotency
checks for existing successful transactions, if found it returns the result of the transaction, instead of fault double entry in case network retry or events like that,

## DeadLock prevention
locks wallet in deterministic order, so concurrent transaction must acquire locks on multiple resources in the same sequence, which prevent circular wait condition.

## DeadLock Detection
Postgresql can detect deadlock when locking rows for update, which can help early rollback and exit, there are retry mechanism in service layer.

## insufficient balance transaction prevention
this function also checks if user have sufficient balance to perform the transaction.

## creates transaction record and double entry in ledger entries
double entry makes the system fully auditable, transaction record duplicate transaction in case of retries, since this whole function is inside of a transaction every statement must be fully executed or roll backed to previous state.

## cache balance in wallet
cached balance can be fetched in an instant, instead of coalesce whole transactions credits - debits to find current balance, for that balance is always updated inside the function in a single transaction for preventing sync issues.

------------------------------------------------------------------------


## Concurrency Safety


The service handles:


-   row‑level locking
-   deadlock retries (service layer)
-   atomic balance updates
-   double‑entry ledger consistency


Designed to scale toward **millions of users**.


