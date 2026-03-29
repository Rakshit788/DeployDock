FROM golang:1.26-alpine AS builder
WORKDIR /app

# Copy the workspace file
COPY go.work go.work.sum* ./

# YOU MUST COPY ALL MOD FILES MENTIONED IN go.work
COPY apps/api/go.mod apps/api/go.sum* ./apps/api/
COPY apps/worker/go.mod apps/worker/go.sum* ./apps/worker/
COPY apps/deployer/go.mod apps/deployer/go.sum* ./apps/deployer/
COPY packages/go.mod packages/go.sum* ./packages/

# Now it will work because the workspace is satisfied
RUN go mod download

COPY . .
WORKDIR /app/apps/worker 
RUN go build -o main ./cmd/worker

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/apps/worker/main .
CMD ["./main"]