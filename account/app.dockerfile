FROM golang:1.21-alpine AS build

RUN apk add --no-cache gcc g++ make ca-certificates
WORKDIR /go/src/github.com/supersection/go-graphql-microservice

COPY go.mod go.sum ./
RUN go mod download

COPY account account
RUN go build -o /go/bin/app ./account/cmd/account


FROM alpine:3.18
WORKDIR /usr/bin
COPY --from=build /go/bin/app .
EXPOSE 8080
CMD ["./app"]
