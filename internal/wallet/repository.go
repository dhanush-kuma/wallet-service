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
			return 0,fmt.Errorf("wallet with id %s not found", walletId)
		}
		return 0,err
	}

	return balance,nil
}