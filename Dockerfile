FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Install build dependencies for go-sqlite3
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o main .

# Final stage
FROM alpine:3.19

# Install runtime dependencies for go-sqlite3 and TLS
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

WORKDIR /app

# Copy the binary
COPY --from=builder /app/main .

# Copy static files, templates, data, and assets
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/data ./data
COPY --from=builder /app/assets ./assets

# Create directory for SQLite database
RUN mkdir -p /app/db

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080
CMD ["./main"]
