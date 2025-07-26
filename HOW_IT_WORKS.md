# Loyalty Platform - How It Works

## Overview

The Loyalty Platform is a multi-tenant, event-driven microservices architecture designed for point-of-sale businesses to manage branded loyalty programs. It supports point accrual, stamp cards, event-based rewards, tiering, and gamification features.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Point of Sale │    │   Admin Portal  │    │   Mobile App    │
│   (External)    │    │   (Future)      │    │   (Future)      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      Apache Kafka         │
                    │   Event Stream Platform   │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐    ┌───────────▼──────────┐    ┌────────▼────────┐
│  Stream        │    │   Analytics          │    │   Marketing     │
│  Processor     │    │   Services           │    │   Engine        │
│  (Real-time)   │    │   (Batch)            │    │   (Future)      │
└───────┬────────┘    └───────────┬──────────┘    └─────────────────┘
        │                         │
        │                         │
┌───────▼────────┐    ┌───────────▼──────────┐
│  Ledger        │    │   MongoDB            │
│  Service       │    │   (Analytics Data)   │
│  (TigerBeetle) │    └──────────────────────┘
└───────┬────────┘
        │
┌───────▼────────┐
│  Membership    │
│  Service       │
│  (MongoDB)     │
└─────────────────┘
```

## Core Services

### 1. Ledger Service (Port 8001)
**Purpose**: Financial accounting and point/stamp management using double-entry bookkeeping

**Technology Stack**:
- **Database**: TigerBeetle (currently mocked for development)
- **Framework**: Go with Gin HTTP server
- **Key Features**: Atomic transactions, audit trail, multi-tenant accounting

**Key Endpoints**:
- `POST /api/v1/accounts` - Create customer loyalty accounts
- `POST /api/v1/transfers` - Process point/stamp transactions
- `GET /api/v1/balance` - Get customer balance
- `GET /api/v1/accounts/:id` - Get account details

**Data Model**:
```go
type Account struct {
    ID             string      // Unique account identifier
    OrgID          string      // Organization ID (multi-tenant)
    CustomerID     string      // Customer identifier
    AccountType    AccountType // Asset, Liability, Equity, etc.
    Code           uint16      // Account code
    DebitsPosted   uint64      // Posted debits
    CreditsPosted  uint64      // Posted credits
    Timestamp      uint64      // Last update timestamp
}
```

### 2. Membership Service (Port 8002)
**Purpose**: Customer and organization management

**Technology Stack**:
- **Database**: MongoDB
- **Framework**: Go with Gin HTTP server
- **Key Features**: Customer profiles, organization management, location settings

**Key Endpoints**:
- `POST /api/v1/customers` - Create customer profiles
- `GET /api/v1/customers/:id` - Get customer details
- `GET /api/v1/customers` - List customers by organization
- `POST /api/v1/organizations` - Create organizations
- `POST /api/v1/locations` - Manage store locations

**Data Model**:
```go
type Customer struct {
    ID           primitive.ObjectID
    CustomerID   string
    OrgID        string
    Email        string
    Phone        string
    FirstName    string
    LastName     string
    DateOfBirth  *time.Time
    Address      Address
    Preferences  CustomerPrefs
    Tier         string
    Status       string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 3. Stream Processor
**Purpose**: Real-time event processing from Kafka streams

**Technology Stack**:
- **Message Queue**: Apache Kafka
- **Language**: Go
- **Key Features**: Event-driven processing, fault tolerance, horizontal scaling

**Event Types Processed**:
- `*.pos.transaction` - Point-of-sale transactions
- `*.loyalty.action` - Manual loyalty actions
- `*.customer.updated` - Customer profile updates

**Processing Flow**:
1. Consumes events from Kafka topics
2. Validates event structure and data
3. Calls appropriate services (Ledger/Membership)
4. Records processing results
5. Handles errors and retries

### 4. Analytics Services

#### RFM Processor
**Purpose**: Calculate Recency, Frequency, Monetary (RFM) scores for customer segmentation

**Features**:
- Real-time RFM calculation
- Customer activity tracking
- Transaction history analysis
- Segmentation data for marketing

#### Tier Processor
**Purpose**: Manage customer loyalty tiers based on spending and visit patterns

**Features**:
- Automatic tier upgrades/downgrades
- Spending threshold monitoring
- Visit frequency tracking
- Scheduled recalculation

## Event Flow Architecture

### 1. Event Production
Events are produced by external systems (POS, admin tools) and follow this structure:

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
    "items": [...],
    "payment_method": "credit_card"
  }
}
```

### 2. Event Routing
Events are routed to Kafka topics using the pattern: `<orgId>.<service>.<event_type>`

Examples:
- `brand123.pos.transaction`
- `brand123.loyalty.action`
- `brand123.customer.updated`

### 3. Event Processing Pipeline

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   External  │───▶│   Kafka     │───▶│   Stream    │───▶│   Ledger    │
│   System    │    │   Topics    │    │  Processor  │    │  Service    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │  TigerBeetle│
                                                   │   Database  │
                                                   └─────────────┘
```

### 4. Data Flow Example: POS Transaction

1. **Customer makes purchase** at POS system
2. **POS system publishes event** to `brand123.pos.transaction` topic
3. **Stream processor consumes** the event
4. **Processor validates** event data and customer existence
5. **Processor calculates** points based on transaction amount
6. **Processor calls Ledger service** to create point transfer
7. **TigerBeetle records** double-entry transaction
8. **Analytics processors** update RFM and tier calculations
9. **Processing result** is logged for monitoring

## Financial Integrity

### Double-Entry Accounting
The platform uses TigerBeetle's double-entry accounting system to ensure financial integrity:

- **Each transaction** creates balanced debits and credits
- **Account types**: Asset, Liability, Equity, Revenue, Expense
- **Atomic operations** prevent partial updates
- **Audit trail** for all financial transactions

### Example Point Transaction
When a customer earns 100 points on a $50 purchase:

```
Debit:  Customer Points Account (Asset)     +100 points
Credit: Points Liability Account (Liability) +100 points
```

## Multi-Tenancy

### Organization Isolation
- Each organization has separate accounts in TigerBeetle
- Customer data is scoped by `org_id`
- Event topics are namespaced by organization
- Analytics data is partitioned by organization

### Data Segregation
```go
// All queries include org_id filter
func (r *MongoRepo) GetCustomersByOrg(ctx context.Context, orgID string) ([]models.Customer, error) {
    filter := bson.M{"org_id": orgID}
    // ... query implementation
}
```

## How to Run the Platform

### Prerequisites

Before running the platform, ensure you have the following installed:

- **Docker Desktop** (for infrastructure services)
- **Go 1.21+** (for building services)
- **Apache Kafka** (for event streaming)
- **Git** (for cloning the repository)

### Quick Start (Recommended)

#### 1. Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd Loyalty

# Create bin directory for compiled services
make bin
```

#### 2. Start Infrastructure Services
```bash
# Start MongoDB, Redis, and other infrastructure
docker-compose up -d mongodb redis

# Verify services are running
docker-compose ps
```

#### 3. Build All Services
```bash
# Build all backend services
make build-backend

# Verify binaries were created
ls -la bin/
```

#### 4. Start Core Services
Open multiple terminal windows and run each service:

**Terminal 1 - Ledger Service:**
```bash
cd services/ledger
go run cmd/server/main.go
# Or use the binary: ../../bin/ledger
```

**Terminal 2 - Membership Service:**
```bash
cd services/membership
go run cmd/server/main.go
# Or use the binary: ../../bin/membership
```

**Terminal 3 - Stream Processor:**
```bash
cd services/stream
go run cmd/processor/main.go
# Or use the binary: ../../bin/stream
```

#### 5. Start Analytics Services (Optional)
**Terminal 4 - RFM Processor:**
```bash
cd services/analytics
go run cmd/rfm-processor/main.go
# Or use the binary: ../../bin/rfm-processor
```

**Terminal 5 - Tier Processor:**
```bash
cd services/analytics
go run cmd/tier-processor/main.go
# Or use the binary: ../../bin/tier-processor
```

### Alternative: Using Docker Compose

For a fully containerized setup:

```bash
# Start all services (including application services)
docker-compose up

# Or start specific services
docker-compose up ledger membership stream

# View logs
docker-compose logs -f stream
```

### Testing the Platform

#### 1. Test API Endpoints
```bash
# Test Ledger Service
curl -X POST http://localhost:8001/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "brand123",
    "customer_id": "cust_001",
    "account_type": 1
  }'

# Test Membership Service
curl -X POST http://localhost:8002/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "brand123",
    "email": "test@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### 2. Generate Test Events
```bash
# Build the Kafka CLI tool
make build-tools

# Generate POS transactions
./bin/kafka-cli pos --count 5 --interval 2s

# Generate loyalty actions
./bin/kafka-cli loyalty --count 3

# Start continuous event stream
./bin/kafka-cli stream --interval 1s
```

#### 3. Run Test Scripts
```bash
# Test API endpoints
./test-apis.sh

# Test analytics pipeline
./test-analytics.sh

# Run complete system test
./run-complete-test.sh
```

### Development Workflow

#### 1. Local Development
```bash
# Start only infrastructure
docker-compose up mongodb redis

# Build and run services in development mode
make build-backend
./bin/ledger &
./bin/membership &
./bin/stream &
./bin/rfm-processor &
./bin/tier-processor &

# Generate test data
./bin/kafka-cli stream --interval 500ms
```

#### 2. Hot Reloading (Optional)
For development with automatic restarts:

```bash
# Install air for hot reloading
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration files in each service directory
# Then run services with air
cd services/ledger && air
cd services/membership && air
cd services/stream && air
```

#### 3. Debugging
```bash
# Run services with debug logging
DEBUG=true ./bin/ledger
DEBUG=true ./bin/membership
DEBUG=true ./bin/stream

# View service logs
docker-compose logs -f mongodb
docker-compose logs -f redis
```

### Environment Configuration

#### Environment Variables
Create a `.env` file in the root directory:

```bash
# Kafka Configuration
KAFKA_BROKERS=localhost:9092

# Database URLs
MONGO_URL=mongodb://admin:password@localhost:27017/loyalty?authSource=admin
REDIS_URL=redis://localhost:6379

# Service URLs
LEDGER_URL=http://localhost:8001
MEMBERSHIP_URL=http://localhost:8002

# Consumer Groups
CONSUMER_GROUP_ID=loyalty-stream-processor

# Debug Mode
DEBUG=false
```

#### Service-Specific Configuration

**Ledger Service:**
```bash
export TIGERBEETLE_ADDRESS=localhost:8000
export REDIS_URL=redis://localhost:6379
export PORT=8001
```

**Membership Service:**
```bash
export MONGO_URL=mongodb://admin:password@localhost:27017/loyalty?authSource=admin
export REDIS_URL=redis://localhost:6379
export PORT=8002
```

**Stream Processor:**
```bash
export KAFKA_BROKERS=localhost:9092
export LEDGER_URL=http://localhost:8001
export MEMBERSHIP_URL=http://localhost:8002
export CONSUMER_GROUP_ID=loyalty-stream-processor
```

### Troubleshooting

#### Common Issues

**1. Port Already in Use:**
```bash
# Check what's using the port
lsof -i :8001
lsof -i :8002

# Kill the process
kill -9 <PID>
```

**2. Database Connection Issues:**
```bash
# Check MongoDB status
docker-compose ps mongodb

# Restart MongoDB
docker-compose restart mongodb

# Check MongoDB logs
docker-compose logs mongodb
```

**3. Kafka Connection Issues:**
```bash
# Ensure Kafka is running
# If using local Kafka installation:
kafka-topics.sh --list --bootstrap-server localhost:9092

# Create required topics
kafka-topics.sh --create --topic brand123.pos.transaction --bootstrap-server localhost:9092
kafka-topics.sh --create --topic brand123.loyalty.action --bootstrap-server localhost:9092
kafka-topics.sh --create --topic brand123.customer.updated --bootstrap-server localhost:9092
```

**4. Service Build Issues:**
```bash
# Clean and rebuild
make clean
make build-backend

# Check Go modules
go mod tidy
go mod download
```

#### Health Checks
```bash
# Check service health
curl http://localhost:8001/api/v1/health
curl http://localhost:8002/api/v1/health

# Check database connectivity
docker exec -it loyalty-mongodb-1 mongosh --eval "db.runCommand('ping')"
```

### Production Deployment

#### 1. Build Production Images
```bash
# Build all Docker images
docker-compose build

# Tag images for registry
docker tag loyalty-ledger:latest your-registry/loyalty-ledger:latest
docker tag loyalty-membership:latest your-registry/loyalty-membership:latest
docker tag loyalty-stream:latest your-registry/loyalty-stream:latest
```

#### 2. Deploy with Kubernetes
```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods
kubectl get services
```

#### 3. Monitor Services
```bash
# View service logs
kubectl logs -f deployment/ledger-service
kubectl logs -f deployment/membership-service
kubectl logs -f deployment/stream-processor

# Check metrics
kubectl top pods
```

### Development and Testing

### Testing Tools

#### Kafka CLI Tool
Generates test events for development and testing:

```bash
# Generate POS transactions
./bin/kafka-cli pos --count 10 --interval 2s

# Generate loyalty actions
./bin/kafka-cli loyalty --count 5

# Continuous event stream
./bin/kafka-cli stream --interval 1s
```

#### Test Scripts
- `test-apis.sh` - API endpoint testing
- `test-analytics.sh` - Analytics pipeline testing
- `test-system.sh` - End-to-end system testing

## Scalability and Performance

### Horizontal Scaling
- **Stateless services** can be scaled horizontally
- **Kafka partitioning** enables parallel processing
- **Consumer groups** allow multiple processors
- **Database sharding** by organization

### Performance Optimizations
- **Redis caching** for frequently accessed data
- **Batch processing** for analytics calculations
- **Connection pooling** for database connections
- **Async processing** for non-critical operations

## Monitoring and Observability

### Logging
- Structured logging across all services
- Event processing results logged
- Error tracking and alerting
- Performance metrics collection

### Health Checks
- Service health endpoints
- Database connectivity monitoring
- Kafka consumer lag monitoring
- Dependency health checks

## Security Considerations

### Data Protection
- Customer data encryption at rest
- Secure communication between services
- API authentication and authorization
- Audit logging for sensitive operations

### Multi-Tenant Security
- Organization-level data isolation
- Cross-tenant access prevention
- Secure event routing
- Tenant-specific configurations

## Future Enhancements

### Planned Features
- **Marketing Engine**: Campaign management and automation
- **React Admin Dashboard**: Web-based administration interface
- **Flutter Mobile App**: Customer-facing mobile application
- **ClickHouse Analytics**: Advanced analytics and reporting
- **Gamification Engine**: Points challenges and leaderboards

### Technical Improvements
- **Flink Integration**: Advanced stream processing
- **GraphQL API**: Flexible data querying
- **WebSocket Support**: Real-time updates
- **Machine Learning**: Predictive analytics and recommendations

## Deployment Architecture

### Production Setup
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   API Gateway   │    │   Service Mesh  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
┌─────────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
│   Ledger        │    │   Membership    │    │   Stream        │
│   Service       │    │   Service       │    │  Processors     │
│   (Replicas)    │    │   (Replicas)    │    │   (Replicas)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      Infrastructure       │
                    │   (Kafka, MongoDB,        │
                    │    TigerBeetle, Redis)    │
                    └───────────────────────────┘
```

This architecture provides a robust, scalable foundation for loyalty program management with real-time processing capabilities, financial integrity, and multi-tenant support. 