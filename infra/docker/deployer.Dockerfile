# Stage 1: Build the binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 1. Copy the workspace configuration
# The * makes the .sum file optional in case it hasn't been generated yet
COPY go.work go.work.sum* ./

# 2. Copy the module files for ALL apps in the workspace
# This satisfies the 'go.work' requirements so 'go mod download' doesn't fail
COPY apps/api/go.mod apps/api/go.sum* ./apps/api/
COPY apps/worker/go.mod apps/worker/go.sum* ./apps/worker/
COPY apps/deployer/go.mod apps/deployer/go.sum* ./apps/deployer/
COPY packages/go.mod packages/go.sum* ./packages/

# 3. Download dependencies for the entire workspace
RUN go mod download

# 4. Copy the actual source code for everything
COPY . .

# 5. Build the specific Deployer binary
WORKDIR /app/apps/deployer
RUN go build -o main ./cmd/deployer

# Stage 2: Final Runtime Image
FROM alpine:latest
WORKDIR /app

# Install certificates (Essential for a deployer to talk to GitHub/Cloud APIs)
RUN apk add --no-cache ca-certificates

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/apps/deployer/main .

# Start the application
CMD ["./main"]