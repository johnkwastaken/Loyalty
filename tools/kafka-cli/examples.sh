#!/bin/bash

# Kafka CLI Examples for Loyalty Platform Testing

echo "🚀 Loyalty Platform Kafka CLI Examples"
echo "====================================="

# Build the CLI tool
echo "📦 Building kafka-cli..."
go build -o kafka-cli

echo ""
echo "📝 Basic Commands:"
echo ""

echo "1️⃣  Single POS transaction:"
echo "./kafka-cli pos"
echo ""

echo "2️⃣  Multiple transactions for specific customer:"
echo "./kafka-cli pos --customer cust_123 --count 5 --interval 1s"
echo ""

echo "3️⃣  Generate loyalty actions:"
echo "./kafka-cli loyalty --count 3"
echo ""

echo "4️⃣  Customer profile updates:"
echo "./kafka-cli customer --customer cust_123"
echo ""

echo "5️⃣  Continuous event stream:"
echo "./kafka-cli stream --interval 500ms"
echo ""

echo "6️⃣  Performance benchmark:"
echo "./kafka-cli benchmark --count 1000"
echo ""

echo "🎯 Testing Scenarios:"
echo ""

echo "Customer Journey Simulation:"
echo "# New customer makes purchases"
echo "./kafka-cli pos --customer cust_new --count 10 --interval 2s"
echo ""
echo "# Customer gets birthday bonus"
echo "./kafka-cli loyalty --customer cust_new --count 1"
echo ""
echo "# Customer profile updated"
echo "./kafka-cli customer --customer cust_new"
echo ""

echo "Multi-Store Simulation:"
echo "# Store 1 transactions"
echo "./kafka-cli pos --location store001 --count 20"
echo ""
echo "# Store 2 transactions"  
echo "./kafka-cli pos --location store002 --count 15"
echo ""

echo "High-Volume Testing:"
echo "# Load test with 10K events"
echo "./kafka-cli benchmark --count 10000"
echo ""

echo "🔄 Continuous Testing:"
echo "# Run this in background for ongoing testing"
echo "./kafka-cli stream --interval 200ms &"
echo ""

echo "🛑 Stop continuous testing:"
echo "# Press Ctrl+C or kill the background process"
echo "pkill kafka-cli"
echo ""

echo "✅ Ready to test! Start with: ./kafka-cli pos"