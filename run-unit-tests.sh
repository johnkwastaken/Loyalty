#!/bin/bash

# Unit Test Runner for Loyalty Platform
# This script runs all unit tests across all services

set -e

echo "🧪 Starting Unit Test Suite for Loyalty Platform"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests for a service
run_service_tests() {
    local service_name=$1
    local service_path=$2
    
    echo -e "\n${BLUE}Testing ${service_name}...${NC}"
    echo "----------------------------------------"
    
    if [ ! -d "$service_path" ]; then
        echo -e "${YELLOW}Warning: ${service_path} not found, skipping...${NC}"
        return
    fi
    
    cd "$service_path"
    
    # Check if there are any test files
    if ! find . -name "*_test.go" -type f | grep -q .; then
        echo -e "${YELLOW}No test files found in ${service_name}${NC}"
        cd - > /dev/null
        return
    fi
    
    # Run tests with verbose output and coverage
    echo "Running tests with coverage..."
    if go test -v -cover ./... 2>&1; then
        echo -e "${GREEN}✅ ${service_name} tests PASSED${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ ${service_name} tests FAILED${NC}"
        ((FAILED_TESTS++))
    fi
    
    ((TOTAL_TESTS++))
    cd - > /dev/null
}

# Function to run tests for a specific package
run_package_tests() {
    local package_name=$1
    local package_path=$2
    
    echo -e "\n${BLUE}Testing ${package_name}...${NC}"
    echo "----------------------------------------"
    
    if [ ! -d "$package_path" ]; then
        echo -e "${YELLOW}Warning: ${package_path} not found, skipping...${NC}"
        return
    fi
    
    cd "$package_path"
    
    # Check if there are any test files
    if ! find . -name "*_test.go" -type f | grep -q .; then
        echo -e "${YELLOW}No test files found in ${package_name}${NC}"
        cd - > /dev/null
        return
    fi
    
    # Run tests with verbose output and coverage
    echo "Running tests with coverage..."
    if go test -v -cover ./... 2>&1; then
        echo -e "${GREEN}✅ ${package_name} tests PASSED${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ ${package_name} tests FAILED${NC}"
        ((FAILED_TESTS++))
    fi
    
    ((TOTAL_TESTS++))
    cd - > /dev/null
}

# Start from the project root
cd "$(dirname "$0")"

echo "📁 Project root: $(pwd)"

# Test Ledger Service
run_service_tests "Ledger Service" "services/ledger"

# Test Membership Service
run_service_tests "Membership Service" "services/membership"

# Test Stream Service
run_service_tests "Stream Service" "services/stream"

# Test Analytics Service - RFM Calculator
run_package_tests "RFM Calculator" "services/analytics/internal/rfm"

# Test Analytics Service - Tier Calculator
run_package_tests "Tier Calculator" "services/analytics/internal/tiers"

# Test Tools
echo -e "\n${BLUE}Testing Tools...${NC}"
echo "----------------------------------------"

# Test Kafka CLI
if [ -d "tools/kafka-cli" ]; then
    cd tools/kafka-cli
    if find . -name "*_test.go" -type f | grep -q .; then
        echo "Testing Kafka CLI..."
        if go test -v ./... 2>&1; then
            echo -e "${GREEN}✅ Kafka CLI tests PASSED${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}❌ Kafka CLI tests FAILED${NC}"
            ((FAILED_TESTS++))
        fi
        ((TOTAL_TESTS++))
    else
        echo -e "${YELLOW}No test files found in Kafka CLI${NC}"
    fi
    cd - > /dev/null
fi

# Test Mock Kafka
if [ -d "tools/mock-kafka" ]; then
    cd tools/mock-kafka
    if find . -name "*_test.go" -type f | grep -q .; then
        echo "Testing Mock Kafka..."
        if go test -v ./... 2>&1; then
            echo -e "${GREEN}✅ Mock Kafka tests PASSED${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}❌ Mock Kafka tests FAILED${NC}"
            ((FAILED_TESTS++))
        fi
        ((TOTAL_TESTS++))
    else
        echo -e "${YELLOW}No test files found in Mock Kafka${NC}"
    fi
    cd - > /dev/null
fi

# Run integration tests if they exist
echo -e "\n${BLUE}Running Integration Tests...${NC}"
echo "----------------------------------------"

if [ -f "test-apis.sh" ]; then
    echo "Running API integration tests..."
    if ./test-apis.sh > /dev/null 2>&1; then
        echo -e "${GREEN}✅ API integration tests PASSED${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ API integration tests FAILED${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
else
    echo -e "${YELLOW}API integration test script not found${NC}"
fi

# Summary
echo -e "\n${BLUE}Test Summary${NC}"
echo "=================================================="
echo -e "Total test suites: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests PASSED!${NC}"
    exit 0
else
    echo -e "\n${RED}💥 Some tests FAILED!${NC}"
    exit 1
fi 