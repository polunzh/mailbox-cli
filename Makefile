.PHONY: build clean test install run lint fmt check

BINARY_NAME=mailbox
BUILD_DIR=bin
GO=go
CGO_ENABLED=0

# Build the binary to bin/ directory
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application (build first)
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./...

# Run tests with verbose output
test-v:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -v ./...

# Install to $GOPATH/bin or $HOME/go/bin
install:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build artifacts"

# Format code
fmt:
	$(GO) fmt ./...

# Run linter (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Run all checks (fmt, test, build)
check: fmt test build

# Development mode - run with hot reload (requires air)
dev:
	@which air > /dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air -c .air.toml 2>/dev/null || air

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build binary to bin/ directory"
	@echo "  run      - Build and run the application"
	@echo "  test     - Run tests"
	@echo "  test-v   - Run tests with verbose output"
	@echo "  install  - Install to GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  fmt      - Format Go code"
	@echo "  lint     - Run linter"
	@echo "  check    - Run fmt, test, and build"
	@echo "  dev      - Run with hot reload (requires air)"
	@echo "  help     - Show this help"
