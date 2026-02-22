package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"wallet-service/internal/api"
	"wallet-service/internal/db"
	"wallet-service/internal/wallet"
)

func main() {
	// Initialize DB pool (uses DATABASE_URL from env)
	pool, err := db.NewPool()
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}

	// Verify DB connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	log.Println("DB connected!")

	// Wire dependencies
	repo := wallet.NewRepository(pool)
	service := wallet.NewService(repo)
	handler := api.NewHandler(service)

	// Setup router
	r := gin.Default()

	// Wallet routes
	r.GET("/wallets/:wallet_id/balance", handler.GetBalance)

	r.POST("/wallets/:wallet_id/topup", handler.TopUpWallet)

	r.POST("/wallets/:wallet_id/bonus", handler.GrantBonus)

	r.POST("/wallets/:wallet_id/spend", handler.Spend)

	r.POST("/users", handler.CreateUser)

	r.POST("/wallets", handler.CreateWallet)
	
	r.POST("/assets", handler.CreateAsset)

	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}