package models

type Transfer struct {
	ID              string `json:"id"`
	DebitAccountID  string `json:"debit_account_id"`
	CreditAccountID string `json:"credit_account_id"`
	Amount          uint64 `json:"amount"`
	Code            uint16 `json:"code"`
	Reference       string `json:"reference"`
	Timestamp       uint64 `json:"timestamp"`
}

type CreateTransferRequest struct {
	OrgID           string `json:"org_id" binding:"required"`
	CustomerID      string `json:"customer_id"`
	TransactionType string `json:"transaction_type" binding:"required"`
	Amount          uint64 `json:"amount" binding:"required"`
	Code            uint16 `json:"code"`
	Reference       string `json:"reference"`
}

type TransferResponse struct {
	TransferID string `json:"transfer_id"`
	Status     string `json:"status"`
}