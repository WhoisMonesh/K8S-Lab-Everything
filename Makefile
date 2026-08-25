.PHONY: build test clean install install-windows build-all release fmt vet run-demo help

# Binary name
BINARY=cka-lab-runner
BUILD_DIR=bin
CMD_PKG=./cmd/cka-lab-runner

# Cross-compile targets (native on every platform)
PLATFORMS=windows/amd64 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Version info
VERSION?=1.0.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X github.com/WhoisMonesh/K8S-Lab-Everything/internal/update.Version=$(VERSION) -X github.com/WhoisMonesh/K8S-Lab-Everything/internal/update.GitCommit=$(GIT_COMMIT)"

help: ## Show this help
	@echo "CKA Lab Runner - Makefile commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## Build the binary
	@echo "Building $(BINARY) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) -v $(CMD_PKG)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY)"

build-all: ## Cross-compile for windows/linux/macos (amd64 + arm64)
	@echo "Cross-compiling $(BINARY) $(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "  -> $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			$(GOBUILD) $(LDFLAGS) -trimpath \
			-o $(BUILD_DIR)/$(BINARY)-$$os-$$arch$$ext $(CMD_PKG) || exit 1; \
	done
	@echo "All binaries written to $(BUILD_DIR)/"

release: build-all ## Alias for build-all (cross-platform release binaries)

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install: build ## Install binary to /usr/local/bin (Unix only)
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
	@echo "Installation complete!"

install-windows: ## Install binary to %GOPATH%\bin (Windows only; run via: make install-windows)
	@powershell -NoProfile -Command "$$gopath = go env GOPATH; New-Item -ItemType Directory -Force -Path (Join-Path $$gopath 'bin') | Out-Null; Copy-Item '$(BUILD_DIR)/$(BINARY).exe' (Join-Path $$gopath 'bin\cka-lab-runner.exe') -Force; Write-Host ('Installed to ' + (Join-Path $$gopath 'bin\cka-lab-runner.exe'))"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

lint: fmt vet ## Run formatters and linters

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

run-demo: build ## Run the demo script
	@echo "Running demo..."
	@cd /tmp && rm -rf cka-demo && mkdir cka-demo && cd cka-demo && ../$(BUILD_DIR)/$(BINARY) init

dev: build ## Build and show version
	@$(BUILD_DIR)/$(BINARY) --help

ci: lint test build ## Run CI checks (lint, test, build)
	@echo "All CI checks passed!"

.DEFAULT_GOAL := help
