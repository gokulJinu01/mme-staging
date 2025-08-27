# Multi-stage build for Go service
FROM golang:1.24-alpine AS build

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mme-tagging ./cmd/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create necessary directories
RUN mkdir -p /app/logs /app/data && \
    chown -R appuser:appgroup /app && \
    chmod -R g+rwX /app

# Copy binary from build stage
COPY --from=build /app/bin/mme-tagging /app/mme-tagging

# Set ownership and permissions
RUN chown appuser:appgroup /app/mme-tagging && \
    chmod +x /app/mme-tagging

# Switch to non-root user
USER appuser

# Environment variables
# ENV JWT_SECRET=""
# ENV JWT_ISSUER="mme"
# ENV JWT_AUDIENCE="mme"

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/mme-tagging"]