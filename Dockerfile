# Build stage
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copy source code
COPY . .

# Download and tidy dependencies
RUN go mod download && go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gymbot ./cmd/bot

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gymbot .

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the binary
CMD ["./gymbot"]