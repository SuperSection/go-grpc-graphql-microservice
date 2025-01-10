# Use the latest Go image with Alpine as the base for building
FROM golang:latest AS build

# Install necessary build tools and certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory for the build stage
WORKDIR /go/src/github.com/supersection/go-grpc-graphql-microservice

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy only the required source code for the account service
COPY account ./account
COPY catalog ./catalog
COPY order ./order
COPY graphql ./graphql

# Build the account service binary
RUN go build -mod=readonly -o /go/bin/app ./account/cmd/account


# Use a minimal Alpine image for the runtime
FROM alpine:3.18

# Set the working directory for the runtime container
WORKDIR /usr/bin

# Copy the compiled binary from the build stage
COPY --from=build /go/bin/app .

# Expose the service port
EXPOSE 8080

# Run the application as the default command
CMD ["./app"]
