# Build stage
FROM golang:1.24.7-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy backend go mod files
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

# Copy backend source code
WORKDIR /app
COPY backend ./backend

# Build the application
WORKDIR /app/backend
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/offline_me ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/offline_me .
COPY --from=builder /app/frontend/out ./frontend/out

# Expose port
EXPOSE 8080

# Run the application
CMD ["./offline_me"]
