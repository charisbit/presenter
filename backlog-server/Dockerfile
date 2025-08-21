# Simple single-stage Golang build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy source code
COPY . .

# Download dependencies and build
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backlog-mcp-server ./main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary
COPY --from=builder /app/backlog-mcp-server .

# Make it executable
RUN chmod +x backlog-mcp-server

EXPOSE 3001

CMD ["./backlog-mcp-server"]
