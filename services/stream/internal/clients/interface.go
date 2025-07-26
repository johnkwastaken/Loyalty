package clients

// LedgerClientInterface defines the interface for ledger client operations
type LedgerClientInterface interface {
	CreatePointsTransfer(orgID, customerID string, points int, reference string) (*TransferResponse, error)
	CreateStampsTransfer(orgID, customerID string, stamps int, reference string) (*TransferResponse, error)
}

// MembershipClientInterface defines the interface for membership client operations
type MembershipClientInterface interface {
	GetCustomer(customerID string) (*Customer, error)
	GetOrganization(orgID string) (*Organization, error)
} 