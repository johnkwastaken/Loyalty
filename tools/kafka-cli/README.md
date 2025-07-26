# Kafka CLI - Loyalty Platform Event Generator

A command-line tool to generate test events for the loyalty platform Kafka streams.

## Installation

```bash
cd tools/kafka-cli
go build -o kafka-cli
```

## Quick Start

```bash
# Generate a single POS transaction
./kafka-cli pos

# Generate 10 POS transactions for specific customer
./kafka-cli pos --customer cust_123 --count 10

# Generate loyalty actions
./kafka-cli loyalty --count 5

# Start continuous event stream
./kafka-cli stream --interval 500ms

# Run benchmark test
./kafka-cli benchmark --count 10000
```

## Commands

### `pos` - Generate POS Transaction Events
Generates point-of-sale transaction events with realistic product data.

```bash
./kafka-cli pos [flags]

Examples:
  # Single transaction
  ./kafka-cli pos
  
  # Multiple transactions with delay
  ./kafka-cli pos --count 10 --interval 2s
  
  # Specific customer and organization
  ./kafka-cli pos --customer cust_456 --org brand456 --location store002
```

### `loyalty` - Generate Loyalty Action Events
Generates manual loyalty actions like bonus points, stamps, birthday bonuses.

```bash
./kafka-cli loyalty [flags]

Examples:
  # Single loyalty action
  ./kafka-cli loyalty
  
  # Multiple actions
  ./kafka-cli loyalty --count 5 --interval 1s
```

### `customer` - Generate Customer Update Events
Generates customer profile update events for tier changes, preferences.

```bash
./kafka-cli customer [flags]

Examples:
  # Single customer update
  ./kafka-cli customer
  
  # Multiple updates
  ./kafka-cli customer --count 3
```

### `stream` - Continuous Event Stream
Generates a continuous mixed stream of events for testing.

```bash
./kafka-cli stream [flags]

Examples:
  # Default stream (1 event per second)
  ./kafka-cli stream
  
  # High frequency stream
  ./kafka-cli stream --interval 100ms
  
  # Custom organization
  ./kafka-cli stream --org coffee_chain --location downtown
```

### `benchmark` - Performance Testing
Generates high-volume events for performance testing.

```bash
./kafka-cli benchmark [flags]

Examples:
  # 10K events benchmark
  ./kafka-cli benchmark --count 10000
  
  # 100K events to specific org
  ./kafka-cli benchmark --count 100000 --org enterprise_client
```

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--brokers` | `localhost:9092` | Kafka broker addresses |
| `--org` | `brand123` | Organization ID |
| `--location` | `store001` | Location ID |
| `--customer` | (random) | Customer ID |
| `--count` | `1` | Number of events to generate |
| `--interval` | `1s` | Interval between events |

## Event Types Generated

### POS Transaction Events
Topic: `{orgId}.pos.transaction`

```json
{
  "event_id": "evt_1642534567890123",
  "event_type": "pos.transaction",
  "org_id": "brand123",
  "location_id": "store001", 
  "customer_id": "cust_789",
  "timestamp": "2025-01-20T15:30:00Z",
  "payload": {
    "transaction_id": "txn_abc123",
    "amount": 25.50,
    "items": [
      {
        "sku": "COFFEE001",
        "name": "Large Coffee",
        "quantity": 1,
        "unit_price": 4.50,
        "total_price": 4.50,
        "category": "beverages"
      }
    ],
    "payment_method": "credit_card",
    "receipt_number": "RCP123456",
    "cashier": "emp_42"
  }
}
```

### Loyalty Action Events
Topic: `{orgId}.loyalty.action`

```json
{
  "event_id": "evt_1642534567890124",
  "event_type": "loyalty.action",
  "org_id": "brand123",
  "location_id": "store001",
  "customer_id": "cust_789",
  "timestamp": "2025-01-20T15:31:00Z",
  "payload": {
    "action_type": "manual_points",
    "points": 100,
    "stamps": 0,
    "reference": "birthday_bonus",
    "extra_data": {
      "reason": "birthday bonus points",
      "admin": "admin_5"
    }
  }
}
```

### Customer Update Events
Topic: `{orgId}.customer.updated`

```json
{
  "event_id": "evt_1642534567890125",
  "event_type": "customer.updated",
  "org_id": "brand123",
  "location_id": "store001",
  "customer_id": "cust_789",
  "timestamp": "2025-01-20T15:32:00Z",
  "payload": {
    "tier": "gold",
    "preferences": {
      "email_marketing": true,
      "sms_marketing": false,
      "language": "en"
    },
    "updated_at": "2025-01-20T15:32:00Z"
  }
}
```

## Testing Scenarios

### Basic Flow Test
```bash
# 1. Create some customers and transactions
./kafka-cli pos --count 20 --customer cust_001 --interval 500ms

# 2. Add loyalty actions
./kafka-cli loyalty --count 5 --customer cust_001

# 3. Update customer profile
./kafka-cli customer --customer cust_001
```

### Multi-Customer Simulation
```bash
# Generate transactions for multiple customers
for i in {1..10}; do
  ./kafka-cli pos --customer "cust_$(printf "%03d" $i)" --count 5
done
```

### Load Testing
```bash
# High volume test
./kafka-cli benchmark --count 50000 --org load_test
```

### Continuous Testing
```bash
# Run continuous stream for extended testing
./kafka-cli stream --interval 200ms --org stress_test
```

## Environment Variables

You can also set defaults via environment variables:

```bash
export KAFKA_BROKERS="kafka1:9092,kafka2:9092"
export LOYALTY_ORG_ID="my_brand"
export LOYALTY_LOCATION_ID="flagship_store"

./kafka-cli pos
```

## Integration with Services

The generated events will be processed by:
- **Stream Processor** - Points/stamps calculation
- **RFM Processor** - Customer segmentation
- **Tier Processor** - Tier upgrades
- **Analytics Services** - Data aggregation

Monitor the logs of these services to see real-time processing of the generated events.