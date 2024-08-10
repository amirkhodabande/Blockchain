build:
	@go build -o bin/blockchain

run: build
	@./bin/blockchain

test:
	@go test -v ./...

proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto

.PHONY: proto