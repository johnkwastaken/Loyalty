package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MembershipClient struct {
	baseURL    string
	httpClient *http.Client
}

type Customer struct {
	CustomerID  string                 `json:"customer_id"`
	OrgID       string                 `json:"org_id"`
	Email       string                 `json:"email"`
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	Tier        string                 `json:"tier"`
	Status      string                 `json:"status"`
	Preferences CustomerPrefs          `json:"preferences"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type CustomerPrefs struct {
	EmailMarketing bool     `json:"email_marketing"`
	SMSMarketing   bool     `json:"sms_marketing"`
	Categories     []string `json:"categories"`
	Language       string   `json:"language"`
}

type Organization struct {
	OrgID    string      `json:"org_id"`
	Name     string      `json:"name"`
	Settings OrgSettings `json:"settings"`
}

type OrgSettings struct {
	PointsPerDollar    float64           `json:"points_per_dollar"`
	StampsPerVisit     int               `json:"stamps_per_visit"`
	RewardThresholds   []RewardThreshold `json:"reward_thresholds"`
	TierRules          []TierRule        `json:"tier_rules"`
	MaxStampsPerCard   int               `json:"max_stamps_per_card"`
}

type RewardThreshold struct {
	Points      int    `json:"points"`
	Stamps      int    `json:"stamps"`
	RewardType  string `json:"reward_type"`
	RewardValue string `json:"reward_value"`
	Description string `json:"description"`
}

type TierRule struct {
	Name             string  `json:"name"`
	MinSpent         float64 `json:"min_spent"`
	MinVisits        int     `json:"min_visits"`
	PointsMultiplier float64 `json:"points_multiplier"`
	Benefits         []string `json:"benefits"`
}

func NewMembershipClient(baseURL string) *MembershipClient {
	return &MembershipClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *MembershipClient) GetCustomer(customerID string) (*Customer, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/customers/" + customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("customer not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("membership service returned status %d", resp.StatusCode)
	}

	var customer Customer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, fmt.Errorf("failed to decode customer: %w", err)
	}

	return &customer, nil
}

func (c *MembershipClient) GetOrganization(orgID string) (*Organization, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/organizations/" + orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("organization not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("membership service returned status %d", resp.StatusCode)
	}

	var org Organization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		return nil, fmt.Errorf("failed to decode organization: %w", err)
	}

	return &org, nil
}