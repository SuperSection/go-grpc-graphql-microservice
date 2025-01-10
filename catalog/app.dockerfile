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

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy only the required source code for the catalog service
COPY catalog ./catalog

# Build the catalog service binary
RUN go build -o /go/bin/app ./catalog/cmd/catalog


# Use a minimal runtime image
FROM alpine:3.18

# Set the working directory for the runtime stage
WORKDIR /usr/bin

# Copy the compiled binary from the build stage
COPY --from=build /go/bin/app .

# Expose the application port
EXPOSE 8080

# Set the default command to run the application
CMD ["./app"]
