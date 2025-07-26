package repository

import (
	"context"
	"github.com/loyalty/ledger/internal/models"
)

// TigerBeetleRepoInterface defines the interface for TigerBeetle repository operations
type TigerBeetleRepoInterface interface {
	CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error)
	CreateTransfer(ctx context.Context, req *models.CreateTransferRequest) (*models.TransferResponse, error)
	GetAccount(ctx context.Context, accountID string) (*models.Account, error)
	GetBalance(ctx context.Context, orgID, customerID string) (map[string]uint64, error)
	Close() error
} 