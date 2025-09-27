dev-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

gen:
	protoc --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative proto/keeper.proto

tidy:
	go mod tidy

build: tidy
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

lint:
	golangci-lint run ./... --fix

check: build lint 