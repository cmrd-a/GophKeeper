.PHONY: gen mod build run lint check buf-dep test clean

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

gen:
	buf generate
	# sh generate-swagger-ui.sh

mod:
	go mod tidy
	go install tool

build: mod
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

run: build
	@if [ -f .env ]; then \
		echo "Loading environment from .env file..."; \
		set -a; source .env; set +a; \
		bin/server; \
	else \
		echo "No .env file found, running with system environment..."; \
		bin/server; \
	fi

lint:
	golangci-lint run ./... --fix

check: build lint test

test:
	go test -v -race -tags=unit -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

buf-dep:
	buf dep update

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
