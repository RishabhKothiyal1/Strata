# ==============================================================================
# Build Stage
# ==============================================================================
FROM rust:1.88-slim AS builder

WORKDIR /app

# Install pkg-config needed by some crates
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy Cargo configuration first
COPY Cargo.toml ./

# Initialize dummy src to pre-compile and cache dependencies
RUN mkdir src && echo "fn main() {}" > src/main.rs

# Generate Cargo.lock and pre-compile all dependencies (cached layer)
RUN cargo build --release || true

# Copy the actual source code
COPY src ./src

# Rebuild the application using the cached dependencies
RUN touch src/main.rs && cargo build --release

# ==============================================================================
# Run Stage
# ==============================================================================
FROM debian:bookworm-slim

# Install SSL certificates and curl for health check validation
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/target/release/storage .

# Expose the Storage service port
EXPOSE 8084

# Execute the Storage service
CMD ["./storage"]
