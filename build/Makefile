# Lenv Makefile
# Lives in build/. Run from repo root via `make -C build` or from build/ via `make`.
# All go commands run against the repo root (parent of this Makefile).

BINARY_NAME := lenv
BUILD_DIR   := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
REPO_ROOT   := $(abspath $(BUILD_DIR)/..)

VERSION    := $(shell cd $(REPO_ROOT) && git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS    := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

GOCMD   := go
GOBUILD := $(GOCMD) build
GOTEST  := $(GOCMD) test
GOMOD   := $(GOCMD) mod
GOFMT   := $(GOCMD) fmt
GOVET   := $(GOCMD) vet

GO_IN_REPO = cd $(REPO_ROOT) &&

.PHONY: all
all: build

.PHONY: build
build:
	$(GO_IN_REPO) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: build-windows
build-windows:
	$(GO_IN_REPO) GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe .

.PHONY: build-linux
build-linux:
	$(GO_IN_REPO) GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux .

.PHONY: build-darwin
build-darwin:
	$(GO_IN_REPO) GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin .

.PHONY: build-all
build-all: build-windows build-linux build-darwin

.PHONY: test
test:
	$(GO_IN_REPO) $(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage:
	$(GO_IN_REPO) $(GOTEST) -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO_IN_REPO) $(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

.PHONY: fmt
fmt:
	$(GO_IN_REPO) $(GOFMT) ./...

.PHONY: vet
vet:
	$(GO_IN_REPO) $(GOVET) ./...

.PHONY: lint
lint:
	$(GO_IN_REPO) golangci-lint run

.PHONY: tidy
tidy:
	$(GO_IN_REPO) $(GOMOD) tidy

.PHONY: clean
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME).exe \
	      $(BUILD_DIR)/$(BINARY_NAME)-linux $(BUILD_DIR)/$(BINARY_NAME)-darwin \
	      $(BUILD_DIR)/coverage.out $(BUILD_DIR)/coverage.html
	rm -f $(REPO_ROOT)/$(BINARY_NAME) $(REPO_ROOT)/$(BINARY_NAME).exe

.PHONY: install
install:
	$(GO_IN_REPO) $(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

.PHONY: run
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: dev
dev:
	$(GO_IN_REPO) $(GOBUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: help
help:
	@echo "Lenv Build System"
	@echo ""
	@echo "Usage: make -C build [target]   (from repo root)"
	@echo "   or: make [target]            (from build/)"
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
