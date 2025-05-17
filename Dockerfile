FROM golang:1.24 AS build

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-server ./cmd/api

# Use a small image for the runtime
FROM golang:1.24-alpine

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/api-server .

# Expose the port the service runs on
EXPOSE 8080

# Run the application
CMD ["/app/api-server"] 