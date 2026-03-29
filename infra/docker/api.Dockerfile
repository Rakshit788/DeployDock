FROM golang:1.26-alpine AS builder

WORKDIR /app

# 1. Copy the workspace file
COPY go.work go.work.sum* ./

# 2. Copy ALL module files defined in your go.work
# Replace these with the actual paths in your project
COPY apps/api/go.mod apps/api/go.sum* ./apps/api/
COPY apps/worker/go.mod apps/worker/go.sum* ./apps/worker/
COPY apps/deployer/go.mod apps/deployer/go.sum* ./apps/deployer/
COPY packages/go.mod packages/go.sum* ./packages/
# If you have a shared 'packages' folder, copy that too:
# COPY packages/common/go.mod ./packages/common/

# 3. Now download everything
RUN go mod download

# 4. Copy the rest of the source code
COPY . .

# 5. Build the specific app
WORKDIR /app/apps/api
RUN go build -o main ./cmd/server

# Stage 2: Final Image
FROM alpine:latest
WORKDIR /app
# Install CA certificates (needed for HTTPS requests/APIs)
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/apps/api/main .
CMD ["./main"]