.PHONY: build clean install test run help

BINARY_NAME=container-composer
BUILD_DIR=bin
VERSION?=dev
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/firasmosbahi/container-composer/cli.Version=$(VERSION) -X github.com/firasmosbahi/container-composer/cli.BuildDate=$(BUILD_DATE)"

help:
	@echo "Container Composer - Makefile commands:"
	@echo "  make build       - Build the binary"
	@echo "  make install     - Install to GOPATH/bin"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make run         - Build and run"
	@echo "  make help        - Show this help message"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) ./cmd/$(BINARY_NAME)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

lint: fmt vet
	@echo "Linting complete"