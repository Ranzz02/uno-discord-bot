# Use the latest official Golang image as a builder
FROM golang:latest AS builder

# Set the working directory
WORKDIR /bot

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy yhe rest of the application
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bot .

# Use a lightweight image for the final executable
FROM alpine:latest

# Set the working directory
WORKDIR /bot

# Copy the compiled binary from the builder stage
COPY --from=builder /bot/bot .

# Run the application
CMD ["./bot"]