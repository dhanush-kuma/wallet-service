package api

import (
    "net/http"
	"errors"

    "github.com/gin-gonic/gin"
    "wallet-service/internal/wallet"
	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
)

type Handler struct{
	walletService *wallet.Service
}

type TopUpRequest struct {
    ReferenceID string `json:"reference_id" binding:"required"`
    Asset       string `json:"asset" binding:"required"`
    Amount      int64  `json:"amount" binding:"required,gt=0"`
}

type BonusRequest struct {
    ReferenceID string `json:"reference_id" binding:"required"`
    Asset       string `json:"asset" binding:"required"`
    Amount      int64  `json:"amount" binding:"required,gt=0"`
}

type SpendRequest struct {
    ReferenceID string `json:"reference_id" binding:"required"`
    Asset       string `json:"asset" binding:"required"`
    Amount      int64  `json:"amount" binding:"required,gt=0"`
}

func NewHandler(ws *wallet.Service) *Handler{
	return &Handler{walletService: ws}
}

func (h *Handler) GetBalance(c *gin.Context){
	walletIDStr := c.Param("wallet_id")

    walletId, err := uuid.Parse(walletIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "invalid wallet_id",
        })
        return
    }
	
	balance, err := h.walletService.GetBalance(c.Request.Context(), walletId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
            c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet_id": walletId,
		"balance": balance,
	})
}

func (h *Handler) TopUpWallet(c *gin.Context) {
    walletIDStr := c.Param("wallet_id")

    walletID, err := uuid.Parse(walletIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
        return
    }

    var req TopUpRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    asset := wallet.AssetCode(req.Asset)

    err = h.walletService.TopUpUserWallet(
        c.Request.Context(),
        req.ReferenceID,
        walletID,
        asset,
        req.Amount,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
    })
}

func (h *Handler) GrantBonus(c *gin.Context) {
    walletIDStr := c.Param("wallet_id")

    walletID, err := uuid.Parse(walletIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
        return
    }

    var req BonusRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    asset := wallet.AssetCode(req.Asset)

    err = h.walletService.GrantBonus(
        c.Request.Context(),
        req.ReferenceID,
        walletID,
        asset,
        req.Amount,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "bonus granted",
    })
}

func (h *Handler) Spend(c *gin.Context) {
    walletIDStr := c.Param("wallet_id")

    walletID, err := uuid.Parse(walletIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
        return
    }

    var req SpendRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    asset := wallet.AssetCode(req.Asset)

    err = h.walletService.SpendFromWallet(
        c.Request.Context(),
        req.ReferenceID,
        walletID,
        asset,
        req.Amount,
    )

    if err != nil {
        if err.Error() == "insufficient balance" {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "spend successful",
    })
}