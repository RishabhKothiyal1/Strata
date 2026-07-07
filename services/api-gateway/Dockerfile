# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy source code and module files
COPY . .

# Run go mod tidy to resolve dependencies and generate go.sum entries
RUN go mod tidy

# Build a static, optimized Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-gateway .

# ==============================================================================
# Run Stage
# ==============================================================================
FROM alpine:latest

# Install base certificates and curl for health check validation
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/api-gateway .

# Expose the API Gateway port
EXPOSE 8000

# Execute the gateway
CMD ["./api-gateway"]
