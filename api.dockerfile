FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go module files first for better layer caching
COPY ./api/go.mod ./api/go.sum ./
RUN go mod download

# Copy the source code
COPY ./api/ ./

# Copy the .env file to the expected location
COPY .env ../

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o urlshortener ./main.go

# Use a minimal alpine image for the final container
FROM alpine:3.18

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/urlshortener .

# Copy the .env file into the expected location
COPY .env ../

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the port the server listens on
EXPOSE 8080

# Command to run the application
CMD ["./urlshortener"]