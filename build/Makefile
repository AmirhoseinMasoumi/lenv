# Lenv Makefile
# Lives in build/. Run from repo root via `make -C build` or from build/ via `make`.
# Uses relative paths so it works cross-platform without abspath/cd tricks.
#   - cwd is always build/ (make -C sets it, otherwise the user is already there)
#   - source dir is the parent (..)
#   - output goes here (.)

BINARY_NAME := lenv

VERSION    := $(shell git -C .. describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo unknown)
LDFLAGS    := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

GOCMD   := go
GOBUILD := $(GOCMD) build
GOTEST  := $(GOCMD) test
GOMOD   := $(GOCMD) mod
GOFMT   := $(GOCMD) fmt
GOVET   := $(GOCMD) vet

.PHONY: all
all: build

.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) -o ./$(BINARY_NAME) ../

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ./$(BINARY_NAME).exe ../

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ./$(BINARY_NAME)-linux ../

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ./$(BINARY_NAME)-darwin ../

.PHONY: build-all
build-all: build-windows build-linux build-darwin

.PHONY: test
test:
	$(GOTEST) -v ../...

.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=./coverage.out ../...
	$(GOCMD) tool cover -html=./coverage.out -o ./coverage.html

.PHONY: fmt
fmt:
	$(GOFMT) ../...

.PHONY: vet
vet:
	$(GOVET) ../...

.PHONY: tidy
tidy:
	$(GOMOD) -C .. tidy

.PHONY: clean
clean:
	-rm -f ./$(BINARY_NAME) ./$(BINARY_NAME).exe
	-rm -f ./$(BINARY_NAME)-linux ./$(BINARY_NAME)-darwin
	-rm -f ./coverage.out ./coverage.html
	-rm -f ../$(BINARY_NAME) ../$(BINARY_NAME).exe

.PHONY: install
install:
	$(GOBUILD) $(LDFLAGS) -o "$(GOPATH)/bin/$(BINARY_NAME)" ../

.PHONY: dev
dev:
	$(GOBUILD) -race $(LDFLAGS) -o ./$(BINARY_NAME) ../

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
	@echo "  tidy           Tidy go modules"
	@echo "  clean          Remove build artifacts"
	@echo "  install        Install to GOPATH/bin"
	@echo "  dev            Build with race detector"
	@echo "  help           Show this help"
