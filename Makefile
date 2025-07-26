.PHONY: build-backend test clean

# Build all backend services
build-backend:
	cd services/ledger && go build -o ../../bin/ledger ./cmd/server
	cd services/membership && go build -o ../../bin/membership ./cmd/server
	cd services/stream && go build -o ../../bin/stream ./cmd/processor
	cd services/analytics && go build -o ../../bin/rfm-processor ./cmd/rfm-processor
	cd services/analytics && go build -o ../../bin/tier-processor ./cmd/tier-processor

# Run tests for all services
test:
	cd services/ledger && go test ./...
	cd services/membership && go test ./...
	cd services/stream && go test ./...
	cd services/analytics && go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build CLI tools
build-tools:
	cd tools/kafka-cli && go build -o ../../bin/kafka-cli

# Create bin directory
bin:
	mkdir -p bin