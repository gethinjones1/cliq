.PHONY: build build-all test bench clean install run dev lint fmt help

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

# Output binary name
BINARY := cliq

# Go build settings
GO := go
GOFLAGS := -trimpath
CGO_ENABLED := 0

# Default target
all: build

## build: Build the binary for current platform
build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY) .

## build-cgo: Build with CGO enabled (for llama.cpp)
build-cgo:
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY) .

## build-all: Build for all supported platforms
build-all:
	@./scripts/build.sh

## test: Run tests
test:
	$(GO) test -v ./...

## test-cover: Run tests with coverage
test-cover:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## bench: Run benchmarks
bench:
	$(GO) test -bench=. -benchmem ./...

## lint: Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## clean: Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f $(BINARY)-*
	rm -rf dist/
	rm -f coverage.out coverage.html

## install: Install to /usr/local/bin
install: build
	sudo cp $(BINARY) /usr/local/bin/

## install-user: Install to ~/.local/bin
install-user: build
	mkdir -p ~/.local/bin
	cp $(BINARY) ~/.local/bin/

## uninstall: Remove from /usr/local/bin
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY)

## run: Run the application
run: build
	./$(BINARY)

## dev: Run with verbose output
dev:
	$(GO) run . -v

## deps: Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## update-deps: Update dependencies
update-deps:
	$(GO) get -u ./...
	$(GO) mod tidy

## release: Create a release using goreleaser
release:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not installed. Run: go install github.com/goreleaser/goreleaser@latest"; \
	fi

## snapshot: Create a snapshot release (for testing)
snapshot:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed"; \
	fi

## help: Show this help message
help:
	@echo "Cliq - AI-powered CLI assistant for Neovim and tmux"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
