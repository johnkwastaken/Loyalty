# Loyalty Platform

A multi-tenant loyalty platform built for point-of-sale businesses with support for branded loyalty programs, point accrual, stamp cards, event-based rewards, tiering and gamification.

## Architecture

- **Ledger Service**: TigerBeetle wrapper for double-entry accounting of points and stamps
- **Membership Service**: MongoDB-based customer and organization management
- **Stream Processor**: Kafka consumer for real-time event processing
- **Docker Compose**: Local development environment

## Tech Stack

- **Backend**: Go microservices
- **Database**: MongoDB (customer data), TigerBeetle (financial ledger)
- **Messaging**: Apache Kafka
- **Cache**: Redis
- **Containers**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Docker Desktop
- Go 1.21+
- Access to Kafka brokers

### Local Development

1. **Start infrastructure services:**
```bash
docker-compose up mongodb tigerbeetle redis
```

2. **Build backend services:**
```bash
make build-backend
```

3. **Run tests:**
```bash
make test
```

4. **Start services:**
```bash
# Terminal 1 - Ledger Service
cd services/ledger && go run cmd/server/main.go

# Terminal 2 - Membership Service  
cd services/membership && go run cmd/server/main.go

# Terminal 3 - Stream Processor
cd services/stream && go run cmd/processor/main.go
```

### Using Docker Compose

```bash
# Start all services
docker-compose up

# Start specific services
docker-compose up ledger membership
```

## API Endpoints

### Ledger Service (Port 8001)

- `POST /api/v1/accounts` - Create account
- `GET /api/v1/accounts/:id` - Get account
- `POST /api/v1/transfers` - Create transfer
- `GET /api/v1/balance` - Get customer balance
- `GET /api/v1/health` - Health check

### Membership Service (Port 8002)

- `POST /api/v1/customers` - Create customer
- `GET /api/v1/customers/:id` - Get customer
- `GET /api/v1/customers` - List customers by org
- `PATCH /api/v1/customers/:id` - Update customer
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/:id` - Get organization
- `GET /api/v1/health` - Health check

## Event Processing

The stream processor consumes Kafka events following the pattern:
`<orgId>.<service>.<event_type>`

### Supported Events

- `*.pos.transaction` - Point-of-sale transactions
- `*.loyalty.action` - Manual loyalty actions
- `*.customer.updated` - Customer profile updates

### Example POS Transaction Event

```json
{
  "event_id": "evt_123456789",
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
    "receipt_number": "RCP001234",
    "cashier": "emp_456"
  }
}
```

## Environment Variables

### Ledger Service
- `TIGERBEETLE_ADDRESS` - TigerBeetle server address (default: localhost:8000)
- `REDIS_URL` - Redis connection URL
- `PORT` - Service port (default: 8001)

### Membership Service
- `MONGO_URL` - MongoDB connection string
- `REDIS_URL` - Redis connection URL  
- `PORT` - Service port (default: 8002)

### Stream Processor
- `KAFKA_BROKERS` - Comma-separated Kafka broker addresses
- `LEDGER_URL` - Ledger service URL (default: http://localhost:8001)
- `MEMBERSHIP_URL` - Membership service URL (default: http://localhost:8002)
- `CONSUMER_GROUP_ID` - Kafka consumer group (default: loyalty-stream-processor)

## Development Commands

```bash
# Build all backend services
make build-backend

# Run tests
make test

# Clean build artifacts
make clean

# Start TigerBeetle locally
tb run replica --address 0.0.0.0:8000 --disk ./tb-data

# Start Flink (when implemented)
docker-compose up flink
```

## Event Naming Conventions

- Topics follow pattern: `<orgId>.<service>.<event_type>`
- Example: `brand123.pos.transaction`
- Each message includes: orgId, locationId, customerId, timestamp, payload

## Financial Integrity

- All point and stamp transactions use TigerBeetle's double-entry accounting
- Each organization gets separate liability accounts
- All accruals and redemptions are atomic transfers
- Balances are always consistent and auditable

## Next Steps

- [ ] Add ClickHouse analytics service
- [ ] Implement marketing engine with campaigns
- [ ] Build React admin dashboard  
- [ ] Add Flutter mobile app
- [ ] Implement tier calculation logic
- [ ] Add gamification features