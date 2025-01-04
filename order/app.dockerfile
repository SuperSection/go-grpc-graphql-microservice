# Use the latest Go image with Alpine as the base for building
FROM golang:1.21-alpine3.18 AS build

# Install necessary build tools and certificates
RUN apk add --no-cache gcc g++ make ca-certificates

# Set the working directory
WORKDIR /go/src/github.com/supersection/go-graphql-microservice

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY account account
COPY catalog catalog
COPY order order

# Build the application
RUN go build -mod=readonly -o /go/bin/app ./order/cmd/order

# Use a smaller runtime image for the final stage
FROM alpine:3.18

# Install CA certificates for HTTPS support
RUN apk add --no-cache ca-certificates

# Set the working directory for the runtime container
WORKDIR /usr/bin

# Copy the compiled application from the build stage
COPY --from=build /go/bin/app .

# Expose the application's port
EXPOSE 8080

# Set the default command to run the application
CMD ["./app"]
