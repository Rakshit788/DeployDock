FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum* ./

COPY apps/api/go.mod apps/api/go.sum* ./apps/api/
COPY apps/worker/go.mod apps/worker/go.sum* ./apps/worker/
COPY apps/deployer/go.mod apps/deployer/go.sum* ./apps/deployer/
COPY packages/go.mod packages/go.sum* ./packages/

RUN go mod download

COPY . .
WORKDIR /app/apps/worker 

# 🔥 CRITICAL FIX
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/worker

# -------- FINAL --------
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates docker-cli docker-cli-buildx git

COPY --from=builder /app/apps/worker/main .

# 🔥 FORCE EXECUTION PERMISSION
RUN chmod 755 main

CMD ["./main"]