package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/loyalty/ledger/internal/models"
)

type MockTigerBeetleRepo struct {
	accounts  map[string]*models.Account
	transfers map[string]*models.Transfer
}

func NewMockTigerBeetleRepo() *MockTigerBeetleRepo {
	return &MockTigerBeetleRepo{
		accounts:  make(map[string]*models.Account),
		transfers: make(map[string]*models.Transfer),
	}
}

func (r *MockTigerBeetleRepo) CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	accountID := r.generateStringID()
	
	account := &models.Account{
		ID:           accountID,
		OrgID:        req.OrgID,
		CustomerID:   req.CustomerID,
		AccountType:  req.AccountType,
		Code:         req.Code,
		DebitsPosted: 0,
		CreditsPosted: 0,
		Timestamp:    uint64(time.Now().Unix()),
	}

	r.accounts[accountID] = account
	
	log.Printf("Mock: Created account %s for customer %s in org %s", 
		accountID, req.CustomerID, req.OrgID)
	
	return account, nil
}

func (r *MockTigerBeetleRepo) CreateTransfer(ctx context.Context, req *models.CreateTransferRequest) (*models.TransferResponse, error) {
	transferID := r.generateStringID()
	
	// Mock double-entry logic
	debitAccountID := r.generateOrgLiabilityAccount(req.OrgID)
	creditAccountID := r.generateCustomerPointsAccount(req.OrgID, req.CustomerID)
	
	if req.TransactionType == "points_redemption" || req.TransactionType == "stamps_redemption" {
		debitAccountID, creditAccountID = creditAccountID, debitAccountID
	}
	
	transfer := &models.Transfer{
		ID:              transferID,
		DebitAccountID:  debitAccountID,
		CreditAccountID: creditAccountID,
		Amount:          req.Amount,
		Code:            req.Code,
		Reference:       req.Reference,
		Timestamp:       uint64(time.Now().Unix()),
	}

	r.transfers[transferID] = transfer
	
	// Update account balances in mock
	r.updateAccountBalance(debitAccountID, req.Amount, true)
	r.updateAccountBalance(creditAccountID, req.Amount, false)
	
	log.Printf("Mock: Created transfer %s: %s -> %s (%d %s)", 
		transferID, debitAccountID, creditAccountID, req.Amount, req.TransactionType)
	
	return &models.TransferResponse{
		TransferID: transferID,
		Status:     "success",
	}, nil
}

func (r *MockTigerBeetleRepo) GetAccount(ctx context.Context, accountID string) (*models.Account, error) {
	account, exists := r.accounts[accountID]
	if !exists {
		return nil, fmt.Errorf("account not found")
	}
	return account, nil
}

func (r *MockTigerBeetleRepo) GetBalance(ctx context.Context, orgID, customerID string) (map[string]uint64, error) {
	pointsAccountID := r.generateCustomerPointsAccount(orgID, customerID)
	stampsAccountID := r.generateCustomerStampsAccount(orgID, customerID)
	
	balances := map[string]uint64{
		"points": 0,
		"stamps": 0,
	}
	
	if account, exists := r.accounts[pointsAccountID]; exists {
		balances["points"] = account.CreditsPosted - account.DebitsPosted
	}
	
	if account, exists := r.accounts[stampsAccountID]; exists {
		balances["stamps"] = account.CreditsPosted - account.DebitsPosted
	}
	
	return balances, nil
}

func (r *MockTigerBeetleRepo) Close() error {
	log.Println("Mock: TigerBeetle repository closed")
	return nil
}

// Helper methods
func (r *MockTigerBeetleRepo) generateStringID() string {
	var bytes [8]byte
	rand.Read(bytes[:])
	return fmt.Sprintf("%x", bytes)
}

func (r *MockTigerBeetleRepo) generateOrgLiabilityAccount(orgID string) string {
	return fmt.Sprintf("liability_%s", orgID)
}

func (r *MockTigerBeetleRepo) generateCustomerPointsAccount(orgID, customerID string) string {
	return fmt.Sprintf("points_%s_%s", orgID, customerID)
}

func (r *MockTigerBeetleRepo) generateCustomerStampsAccount(orgID, customerID string) string {
	return fmt.Sprintf("stamps_%s_%s", orgID, customerID)
}

func (r *MockTigerBeetleRepo) updateAccountBalance(accountID string, amount uint64, isDebit bool) {
	account, exists := r.accounts[accountID]
	if !exists {
		// Create account if it doesn't exist
		account = &models.Account{
			ID:            accountID,
			DebitsPosted:  0,
			CreditsPosted: 0,
			Timestamp:     uint64(time.Now().Unix()),
		}
		r.accounts[accountID] = account
	}
	
	if isDebit {
		account.DebitsPosted += amount
	} else {
		account.CreditsPosted += amount
	}
}