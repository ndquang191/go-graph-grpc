# Use a recent Golang image with Alpine
FROM golang:1.23.1-alpine3.18 AS build

# Install necessary build dependencies
RUN apk add --no-cache gcc g++ make ca-certificates

# Set the working directory
WORKDIR /go/src/github.com/ndquang191/go-graph-grpc

# Copy dependency files
COPY go.mod go.sum ./

# Download and verify dependencies
RUN go mod download

# Copy application source code
COPY vendor vendor
COPY order order

# Build the application
RUN GO111MODULE=on go build -mod=vendor -o /go/bin/app ./order/cmd/order

# Use a minimal runtime image
FROM alpine:3.18

# Set the working directory
WORKDIR /app

# Copy the built application from the builder stage
COPY --from=build /go/bin/app .

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./app"]
