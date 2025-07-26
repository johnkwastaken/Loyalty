package models

type AccountType uint32

const (
	AccountTypeAsset      AccountType = 1
	AccountTypeLiability  AccountType = 2
	AccountTypeEquity     AccountType = 3
	AccountTypeRevenue    AccountType = 4
	AccountTypeExpense    AccountType = 5
)

type Account struct {
	ID             string      `json:"id"`
	OrgID          string      `json:"org_id"`
	CustomerID     string      `json:"customer_id"`
	AccountType    AccountType `json:"account_type"`
	Code           uint16      `json:"code"`
	DebitsPosted   uint64      `json:"debits_posted"`
	DebitsPending  uint64      `json:"debits_pending"`
	CreditsPosted  uint64      `json:"credits_posted"`
	CreditsPending uint64      `json:"credits_pending"`
	Timestamp      uint64      `json:"timestamp"`
}

type CreateAccountRequest struct {
	OrgID       string      `json:"org_id" binding:"required"`
	CustomerID  string      `json:"customer_id"`
	AccountType AccountType `json:"account_type" binding:"required"`
	Code        uint16      `json:"code"`
}