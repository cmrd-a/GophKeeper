.PHONY: gen mod build run lint check buf-dep
	
include .env
export

gen:
	buf generate
	sh generate-swagger-ui.sh

mod:
	go mod tidy
	go install tool

build: mod
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

run: build
	bin/server

lint:
	golangci-lint run ./... --fix

check: build lint 

buf-dep:
	buf dep update