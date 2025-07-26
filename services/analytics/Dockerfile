FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build RFM processor
RUN go build -o rfm-processor ./cmd/rfm-processor

# Build tier processor  
RUN go build -o tier-processor ./cmd/tier-processor

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/rfm-processor .
COPY --from=builder /app/tier-processor .

# Default to RFM processor
CMD ["./rfm-processor"]