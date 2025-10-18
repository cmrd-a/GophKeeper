.PHONY: gen mod build build-server build-client run run-client lint check buf-dep test test-unit test-integration test-client clean

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

build: mod build-server build-client

build-server: mod
	go build -o bin/server ./cmd/server

build-client: mod
	go build -o bin/gophkeeper-client ./cmd/client

run: build-server
	@if [ -f .env ]; then \
		echo "Loading environment from .env file..."; \
		set -a; source .env; set +a; \
		bin/server; \
	else \
		echo "No .env file found, running with system environment..."; \
		bin/server; \
	fi

run-client: build-client
	./scripts/run-client.sh

lint:
	golangci-lint run ./... --fix

check: build lint test

test:
	go test -v -race -tags=unit -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-unit:
	./scripts/run-tests.sh --unit-only

test-integration:
	./scripts/run-tests.sh --integration-only

test-client:
	./scripts/run-tests.sh

test-client-verbose:
	./scripts/run-tests.sh --verbose

test-client-no-coverage:
	./scripts/run-tests.sh --no-coverage

buf-dep:
	buf dep update

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
