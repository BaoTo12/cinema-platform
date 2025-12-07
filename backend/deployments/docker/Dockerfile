# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/api

# Development stage with Air
FROM golang:1.22-alpine AS development

WORKDIR /app

RUN apk add --no-cache git make

# Install Air for hot reload
RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:3.19 AS production

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/config/config.yaml /app/config/config.yaml
# Copy any other necessary files (migrations, etc.)

EXPOSE 8080

CMD ["/app/server"]
