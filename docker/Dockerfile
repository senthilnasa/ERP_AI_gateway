# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ai-gateway ./cmd/server

# Final Production Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/ai-gateway .
COPY --from=builder /app/config ./config
COPY --from=builder /app/prompts ./prompts

EXPOSE 8080

ENTRYPOINT ["./ai-gateway"]
