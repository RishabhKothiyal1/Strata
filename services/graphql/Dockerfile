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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o graphql .

# ==============================================================================
# Run Stage
# ==============================================================================
FROM alpine:latest

# Install base certificates and curl for health check validation
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/graphql .

# Expose the GraphQL service port
EXPOSE 8087

# Execute the GraphQL engine
CMD ["./graphql"]
