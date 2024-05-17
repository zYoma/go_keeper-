BINARY_NAME=shortener
VERSION=0.0.1

build_server:
	go build -o $(BINARY_NAME) cmd/server/main.go 

run_server: build_server
	./$(BINARY_NAME)

build_client:
	go build -o $(BINARY_NAME) cmd/client/main.go 

run_client: build_client
	./$(BINARY_NAME)

proto:
	protoc \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	proto/keeper.proto