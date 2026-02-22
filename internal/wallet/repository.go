package wallet

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository{
	return &Repository{pool: pool}
}

func (r *Repository) GetWalletBalance(ctx context.Context, walletId uuid.UUID) (int64, error){
	query := `
		SELECT balance FROM wallets WHERE id = $1 
	`
	var balance int64
	err := r.pool.QueryRow(ctx, query, walletId).Scan(&balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return 0, fmt.Errorf("wallet with id %s not found: %w", walletId, pgx.ErrNoRows)
    }
    return 0, fmt.Errorf("database query failed: %w", err)
	}

	return balance,nil
}

func (r *Repository) GetWalletAssetCode(
    ctx context.Context,
    walletID uuid.UUID,
) (string, error) {

    var code string

    err := r.pool.QueryRow(ctx, `
        SELECT a.code
        FROM wallets w
        JOIN assets a ON a.id = w.asset_type_id
        WHERE w.id = $1
    `, walletID).Scan(&code)

    return code, err
}

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

func (r *Repository) CreateUser(
    ctx context.Context,
    id uuid.UUID,
    name string,
) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO users (id, name)
        VALUES ($1, $2)
    `, id, name)

    return err
}

func (r *Repository) CreateAsset(
    ctx context.Context,
    code string,
) (int, error) {

    var id int

    err := r.pool.QueryRow(ctx, `
        INSERT INTO assets (code)
        VALUES ($1)
        RETURNING id
    `, code).Scan(&id)

    return id, err
}

func (r *Repository) CreateWallet(
    ctx context.Context,
    id uuid.UUID,
    label string,
    userID *uuid.UUID,
    assetTypeID int,
) error {

    _, err := r.pool.Exec(ctx, `
        INSERT INTO wallets (
            id,
            label,
            user_id,
            asset_type_id,
            balance
        )
        VALUES ($1, $2, $3, $4, 0)
    `,
        id,
        label,
        userID,
        assetTypeID,
    )

    return err
}

func (r *Repository) ListTransactions(
	ctx context.Context,
	limit int,
	offset int,
) ([]Transaction, error) {

	rows, err := r.pool.Query(ctx, `
		SELECT id, reference_id, type, status, created_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID,
			&t.ReferenceID,
			&t.Type,
			&t.Status,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}

	return txs, rows.Err()
}

func (r *Repository) ListLedgerEntries(
	ctx context.Context,
	limit int,
	offset int,
) ([]LedgerEntry, error) {

	rows, err := r.pool.Query(ctx, `
		SELECT id, transaction_id, wallet_id, direction, amount, created_at
		FROM ledger_entries
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LedgerEntry

	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(
			&e.ID,
			&e.TransactionID,
			&e.WalletID,
			&e.Direction,
			&e.Amount,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}