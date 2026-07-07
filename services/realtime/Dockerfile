# ==============================================================================
# Build Stage
# ==============================================================================
FROM rust:1.85-slim AS builder

WORKDIR /app

# Copy Cargo configuration first
COPY Cargo.toml ./

# Initialize dummy src to pre-compile and cache dependencies
RUN mkdir src && echo "fn main() {}" > src/main.rs

# Generate Cargo.lock by running a dummy update
RUN cargo update

# Downgrade bleeding-edge transitive dependencies to versions compatible with stable Rust 1.80
RUN cargo update -p time --precise 0.3.36
RUN cargo update -p base64ct --precise 1.6.0
RUN cargo update -p url --precise 2.5.0
RUN cargo update -p idna --precise 0.5.0

# Pre-compile dependencies (cached layer)
RUN cargo build --release

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
COPY --from=builder /app/target/release/realtime .

# Expose the Realtime service port
EXPOSE 8083

# Execute the Realtime service
CMD ["./realtime"]
