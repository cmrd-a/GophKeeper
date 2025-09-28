.PHONY: gen mod build lint check buf-dep
gen:
	buf generate
# 	protoc \
# 	--go_out=pb \
# 	--go_opt=paths=source_relative \
# 	--go-grpc_out=pb \
# 	--go-grpc_opt=paths=source_relative \
# 	--grpc-gateway_out=pb \
#     --grpc-gateway_opt paths=source_relative \
#     --grpc-gateway_opt generate_unbound_methods=true \
# 	--openapiv2_out ./pb/openapiv2 \
# 	proto/keeper.proto

mod:
	go mod tidy
	go install tool

build: mod
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

lint:
	golangci-lint run ./... --fix

check: build lint 

buf-dep:
	buf dep update