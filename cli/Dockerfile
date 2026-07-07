FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /usr/local/bin/strata .

FROM alpine:3.19

RUN apk add --no-cache docker docker-compose postgresql-client

COPY --from=builder /usr/local/bin/strata /usr/local/bin/strata

ENTRYPOINT ["strata"]
