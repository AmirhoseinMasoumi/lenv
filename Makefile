# Lenv Makefile
# Build, test, and manage the lenv project

# Variables
BINARY_NAME := lenv
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for Windows
.PHONY: build-windows
build-windows: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe .

# Build for Linux
.PHONY: build-linux
build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux .

# Build for macOS
.PHONY: build-darwin
build-darwin: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin .

# Build all platforms
.PHONY: build-all
build-all: build-windows build-linux build-darwin

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) ./...

# Vet code
.PHONY: vet
vet:
	$(GOVET) ./...

# Lint (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

# Install to GOPATH/bin
.PHONY: install
install:
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

# Run the binary
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Development build with race detector
.PHONY: dev
dev: $(BUILD_DIR)
	$(GOBUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Show help
.PHONY: help
help:
	@echo "Lenv Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build binary for current platform (default)"
	@echo "  build-windows  Build Windows executable"
	@echo "  build-linux    Build Linux binary"
	@echo "  build-darwin   Build macOS binary"
	@echo "  build-all      Build for all platforms"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  fmt            Format code"
	@echo "  vet            Run go vet"
	@echo "  lint           Run golangci-lint"
	@echo "  tidy           Tidy go modules"
	@echo "  clean          Remove build artifacts"
	@echo "  install        Install to GOPATH/bin"
	@echo "  dev            Build with race detector"
	@echo "  help           Show this help"
