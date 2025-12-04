FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the application from cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
# Final Stage
FROM alpine:3.18
WORKDIR /app
# Copy the built binary
COPY --from=builder /server .
# Expose port
EXPOSE 8000

# Run the application
CMD ["./server"]
