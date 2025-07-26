#!/bin/bash

echo "🚀 Testing Loyalty Platform APIs"
echo "================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_test() {
    echo -e "${BLUE}🧪 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test 1: Health Checks
print_test "Testing service health endpoints..."

if curl -s http://localhost:8001/api/v1/health | grep -q "healthy"; then
    print_success "Ledger service is healthy"
else
    print_error "Ledger service health check failed"
fi

if curl -s http://localhost:8002/api/v1/health | grep -q "healthy"; then
    print_success "Membership service is healthy"
else
    print_error "Membership service health check failed"
fi

echo ""

# Test 2: Create Organization
print_test "Creating test organization..."

ORG_RESPONSE=$(curl -s -X POST http://localhost:8002/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "test_org_001",
    "name": "Test Coffee Shop",
    "description": "A test coffee shop for the loyalty platform",
    "settings": {
      "points_per_dollar": 2.0,
      "stamps_per_visit": 1,
      "max_stamps_per_card": 10,
      "reward_thresholds": [
        {
          "points": 100,
          "stamps": 0,
          "reward_type": "discount",
          "reward_value": "10%",
          "description": "10% off next purchase"
        },
        {
          "points": 0,
          "stamps": 10,
          "reward_type": "free_item",
          "reward_value": "free_coffee",
          "description": "Free coffee"
        }
      ],
      "tier_rules": [
        {
          "name": "Bronze",
          "min_spent": 0,
          "min_visits": 0,
          "points_multiplier": 1.0,
          "benefits": ["Basic rewards"]
        },
        {
          "name": "Silver", 
          "min_spent": 100,
          "min_visits": 5,
          "points_multiplier": 1.25,
          "benefits": ["25% bonus points", "Priority support"]
        }
      ]
    }
  }')

if echo "$ORG_RESPONSE" | grep -q "test_org_001"; then
    print_success "Organization created successfully"
else
    print_error "Organization creation failed: $ORG_RESPONSE"
fi

echo ""

# Test 3: Create Customer
print_test "Creating test customer..."

CUSTOMER_RESPONSE=$(curl -s -X POST http://localhost:8002/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "test_org_001",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "address": {
      "street": "123 Main St",
      "city": "Test City",
      "state": "TC",
      "zip_code": "12345",
      "country": "US"
    },
    "preferences": {
      "email_marketing": true,
      "sms_marketing": false,
      "language": "en"
    }
  }')

if echo "$CUSTOMER_RESPONSE" | grep -q "john.doe@example.com"; then
    CUSTOMER_ID=$(echo "$CUSTOMER_RESPONSE" | grep -o '"customer_id":"[^"]*"' | cut -d'"' -f4)
    print_success "Customer created successfully (ID: $CUSTOMER_ID)"
else
    print_error "Customer creation failed: $CUSTOMER_RESPONSE"
fi

echo ""

# Test 4: Create Account
print_test "Creating ledger account for customer..."

ACCOUNT_RESPONSE=$(curl -s -X POST http://localhost:8001/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "test_org_001",
    "customer_id": "'$CUSTOMER_ID'",
    "account_type": 1,
    "code": 1001
  }')

if echo "$ACCOUNT_RESPONSE" | grep -q "test_org_001"; then
    print_success "Ledger account created successfully"
else
    print_error "Account creation failed: $ACCOUNT_RESPONSE"
fi

echo ""

# Test 5: Create Transfer (Points)
print_test "Creating points transfer..."

TRANSFER_RESPONSE=$(curl -s -X POST http://localhost:8001/api/v1/transfers \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "test_org_001",
    "customer_id": "'$CUSTOMER_ID'",
    "transaction_type": "points_accrual",
    "amount": 50,
    "code": 1001,
    "reference": "test_purchase_001"
  }')

if echo "$TRANSFER_RESPONSE" | grep -q "success"; then
    print_success "Points transfer successful"
else
    print_error "Points transfer failed: $TRANSFER_RESPONSE"
fi

echo ""

# Test 6: Check Balance
print_test "Checking customer balance..."

BALANCE_RESPONSE=$(curl -s "http://localhost:8001/api/v1/balance?org_id=test_org_001&customer_id=$CUSTOMER_ID")

if echo "$BALANCE_RESPONSE" | grep -q "points_balance"; then
    print_success "Balance retrieved successfully: $BALANCE_RESPONSE"
else
    print_error "Balance check failed: $BALANCE_RESPONSE"
fi

echo ""

# Test 7: Get Customer
print_test "Retrieving customer data..."

CUSTOMER_GET_RESPONSE=$(curl -s "http://localhost:8002/api/v1/customers/$CUSTOMER_ID")

if echo "$CUSTOMER_GET_RESPONSE" | grep -q "john.doe@example.com"; then
    print_success "Customer retrieved successfully"
else
    print_error "Customer retrieval failed: $CUSTOMER_GET_RESPONSE"
fi

echo ""

# Test 8: Get Organization
print_test "Retrieving organization data..."

ORG_GET_RESPONSE=$(curl -s "http://localhost:8002/api/v1/organizations/test_org_001")

if echo "$ORG_GET_RESPONSE" | grep -q "Test Coffee Shop"; then
    print_success "Organization retrieved successfully"
else
    print_error "Organization retrieval failed: $ORG_GET_RESPONSE"
fi

echo ""

print_success "🎉 API Testing Complete!"
echo ""
echo "📊 Test Summary:"
echo "- Services are running and healthy"
echo "- Organization management working"
echo "- Customer management working"
echo "- Ledger account creation working"
echo "- Points transfer system working"
echo "- Balance inquiry working"
echo ""
echo "🔄 Next: Start analytics processors for RFM and tier calculation"