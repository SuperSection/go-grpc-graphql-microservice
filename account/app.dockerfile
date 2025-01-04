# Use the latest Go image with Alpine as the base for building
FROM golang:1.21-alpine3.18 AS build

# Install necessary build tools and dependencies
RUN apk add --no-cache gcc g++ make ca-certificates

# Set the working directory inside the container
WORKDIR /go/src/github.com/supersection/go-graphql-microservice

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy only the required source code for the account service
COPY account ./account

# Build the account service binary
RUN go build -mod=readonly -o /go/bin/app ./account/cmd/account


# Use a minimal Alpine image for the runtime
FROM alpine:3.18

# Install CA certificates for secure connections
RUN apk add --no-cache ca-certificates

# Set the working directory for the runtime container
WORKDIR /usr/bin

# Copy the compiled binary from the build stage
COPY --from=build /go/bin/app .

# Expose the service port
EXPOSE 8080

# Run the application as the default command
CMD ["./app"]
