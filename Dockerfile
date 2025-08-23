FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o endpoint cmd/server/main.go
RUN go build -o consumer cmd/consumer/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binaries from builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/consumer .

# Copy config files
COPY --from=builder /app/config ./config
COPY --from=builder /app/files ./files

# Default command
CMD ["./server"]