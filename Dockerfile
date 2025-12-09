# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate templ files
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ohabits ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

# Set timezone to Kuwait
ENV TZ=Asia/Kuwait

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ohabits .

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/migrations ./migrations

# Create uploads directory
RUN mkdir -p /app/uploads

# Expose port
EXPOSE 8080

# Run the application
CMD ["./ohabits"]
