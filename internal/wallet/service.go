package wallet

import (
	"context"
	"errors"
	"time"
	"fmt"

	"github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgconn"
)

type Service struct{
	repo *Repository
}

type AssetCode string

const (
    AssetGold    AssetCode = "GOLD"
    AssetDiamond AssetCode = "DIAMOND"
)

var TreasuryWalletByAsset = map[AssetCode]uuid.UUID{
    AssetGold:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
    AssetDiamond: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
}

var RevenueWalletByAsset = map[AssetCode]uuid.UUID{
    AssetGold:    uuid.MustParse("00000000-0000-0000-0000-000000000003"),
    AssetDiamond: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
}

func NewService(repo *Repository) *Service{
	return &Service{repo: repo}
}

func (s *Service) GetBalance(ctx context.Context, walletId uuid.UUID) (int64, error){
	return s.repo.GetWalletBalance(ctx,walletId)
}

func (s *Service) TopUpUserWallet(
    ctx context.Context,
    referenceID string,
    userWalletID uuid.UUID,
    asset AssetCode,
    amount int64,
) error {

    treasuryID, ok := TreasuryWalletByAsset[asset]
    if !ok {
        return errors.New("unsupported asset type")
    }

	    walletAsset, err := s.repo.GetWalletAssetCode(ctx, userWalletID)
    if err != nil {
        return err
    }

    if walletAsset != string(asset) {
        return fmt.Errorf(
            "wallet asset mismatch: wallet=%s request=%s",
            walletAsset,
            asset,
        )
    }

    const maxRetries = 3

    for i := 0; i < maxRetries; i++ {

        err := s.repo.Transfer(
            ctx,
            referenceID,
            treasuryID,
            userWalletID,
            amount,
        )

        if err == nil {
            return nil
        }

        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "40P01" {
            time.Sleep(50 * time.Millisecond)
            continue
        }

        return err
    }

    return errors.New("topup failed after retries")
}

func (s *Service) GrantBonus(
    ctx context.Context,
    referenceID string,
    userWalletID uuid.UUID,
    asset AssetCode,
    amount int64,
) error {

    treasuryID, ok := TreasuryWalletByAsset[asset]
    if !ok {
        return errors.New("unsupported asset type")
    }

    // asset validation
    walletAsset, err := s.repo.GetWalletAssetCode(ctx, userWalletID)
    if err != nil {
        return err
    }

    if walletAsset != string(asset) {
        return fmt.Errorf(
            "wallet asset mismatch: wallet=%s request=%s",
            walletAsset,
            asset,
        )
    }

    const maxRetries = 3

    for i := 0; i < maxRetries; i++ {

        err := s.repo.Transfer(
            ctx,
            referenceID,
            treasuryID,
            userWalletID,
            amount,
        )

        if err == nil {
            return nil
        }

        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "40P01" {
            time.Sleep(50 * time.Millisecond)
            continue
        }

        return err
    }

    return errors.New("bonus grant failed after retries")
}

func (s *Service) SpendFromWallet(
    ctx context.Context,
    referenceID string,
    userWalletID uuid.UUID,
    asset AssetCode,
    amount int64,
) error {

    treasuryID, ok := TreasuryWalletByAsset[asset]
    if !ok {
        return errors.New("unsupported asset type")
    }

    walletAsset, err := s.repo.GetWalletAssetCode(ctx, userWalletID)
    if err != nil {
        return err
    }

    if walletAsset != string(asset) {
        return fmt.Errorf(
            "wallet asset mismatch: wallet=%s request=%s",
            walletAsset,
            asset,
        )
    }

    const maxRetries = 3

    for i := 0; i < maxRetries; i++ {

        err := s.repo.Transfer(
            ctx,
            referenceID,
            userWalletID,
            treasuryID,
            amount,
        )

        if err == nil {
            return nil
        }

        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "40P01" {
            time.Sleep(50 * time.Millisecond)
            continue
        }

        return err
    }

    return errors.New("spend failed after retries")
}