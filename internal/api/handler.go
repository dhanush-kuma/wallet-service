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

        c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
        return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet_id": walletId,
		"balance": balance,
	})
}