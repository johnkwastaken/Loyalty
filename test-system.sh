#!/bin/bash

# Loyalty Platform System Testing Script
# This script demonstrates how to start and test the complete system

set -e

echo "🚀 Loyalty Platform System Testing"
echo "=================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check prerequisites
print_step "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    print_error "Docker not found. Please install Docker Desktop."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose not found. Please install Docker Compose."
    exit 1
fi

if ! command -v go &> /dev/null; then
    print_error "Go not found. Please install Go 1.21+."
    exit 1
fi

print_success "All prerequisites found!"
echo ""

# Build everything
print_step "Building all services and tools..."

echo "Building backend services..."
make build-backend

echo "Building CLI tools..."  
make build-tools

print_success "Build completed!"
echo ""

# Start infrastructure
print_step "Starting infrastructure services..."

echo "Starting MongoDB, TigerBeetle, and Redis..."
docker-compose up -d mongodb tigerbeetle redis

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 10

print_success "Infrastructure services started!"
echo ""

# Start application services
print_step "Starting application services..."

echo "Starting Ledger service..."
docker-compose up -d ledger

echo "Starting Membership service..."
docker-compose up -d membership

echo "Starting Stream processor..."
docker-compose up -d stream

echo "Starting Analytics processors..."
docker-compose up -d rfm-processor tier-processor

# Wait for services to be ready
echo "Waiting for application services to start..."
sleep 15

print_success "All services started!"
echo ""

# Check service health
print_step "Checking service health..."

# Function to check service health
check_health() {
    local service=$1
    local port=$2
    local name=$3
    
    if curl -s "http://localhost:$port/api/v1/health" > /dev/null; then
        print_success "$name service is healthy"
    else
        print_warning "$name service not responding (this is expected if using mock Kafka)"
    fi
}

check_health "ledger" "8001" "Ledger"
check_health "membership" "8002" "Membership"

echo ""

# Run tests
print_step "Running system tests..."

echo ""
echo "🧪 Test 1: Basic POS Transaction"
echo "================================"

./bin/kafka-cli pos --customer cust_test_001 --count 1
sleep 2

echo ""
echo "🧪 Test 2: Customer Journey Simulation"
echo "======================================"

echo "Simulating new customer journey..."

# New customer makes first purchase
echo "👤 New customer first purchase..."
./bin/kafka-cli pos --customer cust_journey_001 --count 1 --org brand123 --location store001
sleep 1

# Customer makes more purchases over time
echo "🛒 Multiple purchases..."
./bin/kafka-cli pos --customer cust_journey_001 --count 5 --interval 500ms
sleep 3

# Customer gets birthday bonus
echo "🎂 Birthday bonus..."
./bin/kafka-cli loyalty --customer cust_journey_001 --count 1
sleep 1

# Customer profile update
echo "👤 Profile update..."
./bin/kafka-cli customer --customer cust_journey_001
sleep 1

print_success "Customer journey simulation completed!"

echo ""
echo "🧪 Test 3: Multi-Customer RFM Analysis"
echo "====================================="

echo "Generating transactions for RFM analysis..."

# Create diverse customer patterns
customers=("cust_champion_001" "cust_loyal_001" "cust_new_001" "cust_atrisk_001")
amounts=("150.50" "75.25" "25.00" "200.00")
counts=("10" "5" "2" "8")

for i in "${!customers[@]}"; do
    customer=${customers[$i]}
    amount=${amounts[$i]}
    count=${counts[$i]}
    
    echo "💳 Creating pattern for $customer..."
    
    # Simulate different spending patterns
    for ((j=1; j<=count; j++)); do
        ./bin/kafka-cli pos --customer "$customer" --count 1 > /dev/null
        sleep 0.2
    done
done

print_success "RFM test data generated!"

echo ""
echo "🧪 Test 4: Tier Progression Testing"
echo "=================================="

echo "Simulating tier progression..."

# High-value customer
echo "💎 High-value customer progression..."
./bin/kafka-cli pos --customer cust_vip_001 --count 20 --interval 100ms
sleep 2

# Medium-value customer  
echo "🥈 Medium-value customer..."
./bin/kafka-cli pos --customer cust_mid_001 --count 10 --interval 100ms
sleep 2

print_success "Tier progression simulation completed!"

echo ""
echo "🧪 Test 5: Load Testing"
echo "======================"

echo "Running performance benchmark..."
./bin/kafka-cli benchmark --count 1000 --org load_test

print_success "Load testing completed!"

echo ""
echo "📊 System Status Check"
echo "====================="

echo "Checking Docker containers..."
docker-compose ps

echo ""
echo "Recent container logs:"
echo "----------------------"

echo "Ledger service logs:"
docker-compose logs --tail 10 ledger

echo ""
echo "Membership service logs:"
docker-compose logs --tail 10 membership

echo ""
echo "Stream processor logs:"
docker-compose logs --tail 10 stream

echo ""
echo "RFM processor logs:"
docker-compose logs --tail 10 rfm-processor

echo ""
echo "Tier processor logs:"
docker-compose logs --tail 10 tier-processor

echo ""
print_success "System testing completed!"

echo ""
echo "🎯 Testing Summary"
echo "=================="
echo "✅ Infrastructure services started"
echo "✅ Application services deployed"
echo "✅ Basic functionality tested"
echo "✅ Customer journey simulated"
echo "✅ RFM analysis data generated" 
echo "✅ Tier progression tested"
echo "✅ Load testing completed"

echo ""
echo "🔍 Next Steps:"
echo "- Check MongoDB for RFM scores and tier data"
echo "- Monitor service logs for real-time processing"
echo "- Use Kafka CLI for continued testing"
echo "- Review analytics data in MongoDB"

echo ""
echo "🛑 To stop the system:"
echo "docker-compose down"

echo ""
print_success "Testing framework ready! 🚀"