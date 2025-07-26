package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/loyalty/ledger/internal/handlers"
	"github.com/loyalty/ledger/internal/repository"
)

func main() {
	log.Println("Starting Ledger Service with Mock TigerBeetle...")
	
	repo := repository.NewMockTigerBeetleRepo()
	defer repo.Close()

	handler := handlers.NewLedgerHandler(repo)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/accounts", handler.CreateAccount)
		v1.GET("/accounts/:id", handler.GetAccount)
		v1.POST("/transfers", handler.CreateTransfer)
		v1.GET("/balance", handler.GetBalance)
		v1.GET("/health", handler.Health)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("Starting ledger service on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}