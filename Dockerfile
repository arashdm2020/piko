# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o piko .

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/piko .
COPY --from=builder /app/config /app/config

# Create data directory for blockchain storage
RUN mkdir -p /app/data

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./piko", "--config", "./config/config.json"] 