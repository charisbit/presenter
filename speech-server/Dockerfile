# Build stage for Go proxy service
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy source code and go.mod
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build the application for the target architecture
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o speech-server ./cmd/main.go

# Final stage with enhanced Japanese TTS support (VOICEVOX + MLX-Audio)
FROM alpine:latest

# Install necessary packages and curl for HTTP requests
RUN apk --no-cache add ca-certificates wget curl

# Set working directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/speech-server .

# Create directories
RUN mkdir -p /app/cache /app/config

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Run the application
CMD ["./speech-server"]