# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o functions .

# ==============================================================================
# Run Stage
# ==============================================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/functions .

EXPOSE 8085

CMD ["./functions"]
