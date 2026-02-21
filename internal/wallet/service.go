package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Service struct{
	repo *Repository
}

func NewService(repo *Repository) *Service{
	return &Service{repo: repo}
}

func (s *Service) GetBalance(ctx context.Context, walletId uuid.UUID) (int64, error){
	return s.repo.GetWalletBalance(ctx,walletId)
}