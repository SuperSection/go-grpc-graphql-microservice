generate:
	cd graphql && go run github.com/99designs/gqlgen generate

account-grpc:
	cd account && protoc --go-grpc_out=./pb --go-grpc_opt=paths=source_relative account.proto

account-pb:
	cd account && protoc --go_out=./pb --go_opt=paths=source_relative account.proto

catalog-grpc:
	cd catalog && protoc --go-grpc_out=./pb --go-grpc_opt=paths=source_relative catalog.proto

catalog-pb:
	cd catalog && protoc --go_out=./pb --go_opt=paths=source_relative catalog.proto

order-grpc:
	cd order && protoc --go-grpc_out=./pb --go-grpc_opt=paths=source_relative order.proto

order-pb:
	cd order && protoc --go_out=./pb --go_opt=paths=source_relative order.proto
