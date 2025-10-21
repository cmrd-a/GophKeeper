.PHONY: gen mod build build-server build-client run lint check buf-dep test demo

# Build variables
VERSION ?= dev
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

gen:
	buf generate

mod:
	go mod tidy
	go install tool

build: mod build-server build-client

build-server: mod
	go build -o bin/server ./cmd/server

build-client: mod
	go build -ldflags "$(LDFLAGS)" -o bin/gophkeeper-client ./cmd/client

run: build-server
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
	go test ./...

buf-dep:
	buf dep update

demo: build
	cp .env.example .env
	docker compose up -d
	./bin/client
