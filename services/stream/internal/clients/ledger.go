package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LedgerClient struct {
	baseURL    string
	httpClient *http.Client
}

type CreateTransferRequest struct {
	OrgID           string `json:"org_id"`
	CustomerID      string `json:"customer_id"`
	TransactionType string `json:"transaction_type"`
	Amount          uint64 `json:"amount"`
	Code            uint16 `json:"code"`
	Reference       string `json:"reference"`
}

type TransferResponse struct {
	TransferID string `json:"transfer_id"`
	Status     string `json:"status"`
}

func NewLedgerClient(baseURL string) *LedgerClient {
	return &LedgerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *LedgerClient) CreatePointsTransfer(orgID, customerID string, points int, reference string) (*TransferResponse, error) {
	req := CreateTransferRequest{
		OrgID:           orgID,
		CustomerID:      customerID,
		TransactionType: "points_accrual",
		Amount:          uint64(points),
		Code:            1,
		Reference:       reference,
	}

	return c.createTransfer(req)
}

func (c *LedgerClient) CreateStampsTransfer(orgID, customerID string, stamps int, reference string) (*TransferResponse, error) {
	req := CreateTransferRequest{
		OrgID:           orgID,
		CustomerID:      customerID,
		TransactionType: "stamps_accrual",
		Amount:          uint64(stamps),
		Code:            2,
		Reference:       reference,
	}

	return c.createTransfer(req)
}

func (c *LedgerClient) createTransfer(req CreateTransferRequest) (*TransferResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/transfers",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ledger service returned status %d", resp.StatusCode)
	}

	var response TransferResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}