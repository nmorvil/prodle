FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o main .

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

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

# Set environment variables
ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

CMD ["./main"]
