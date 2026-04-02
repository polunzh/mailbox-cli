.PHONY: build clean test test-cover coverage install run lint fmt check release-assets release-tag release-push release

BINARY_NAME=mailbox
BUILD_DIR=bin
DIST_DIR=dist
GO=go
CGO_ENABLED=0
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
VERSION ?=
RELEASE_PLATFORMS=linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 freebsd-amd64 openbsd-amd64

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

# Run tests with coverage report
test-cover:
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) test -cover ./... 2>&1 | grep -v "no such tool" || true

# Generate detailed coverage report
coverage:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -coverprofile=$(COVERAGE_FILE) ./...
	@echo "Coverage summary:"
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo ""
	@echo "Generate HTML report: make coverage-html"
	@echo "View HTML: open $(COVERAGE_HTML)"

# Generate HTML coverage report
coverage-html: coverage
	CGO_ENABLED=$(CGO_ENABLED) $(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Generated: $(COVERAGE_HTML)"

# Show coverage for specific package (usage: make coverage-pkg PKG=./tui)
coverage-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make coverage-pkg PKG=./tui"; \
		exit 1; \
	fi
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -cover $(PKG)

# Install to $GOPATH/bin or $HOME/go/bin
install:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Cleaned build artifacts"

release-check:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=v0.1.0"; \
		exit 1; \
	fi
	@if ! printf '%s' "$(VERSION)" | grep -Eq '^v[0-9]'; then \
		echo "VERSION must start with v, e.g. v0.1.0"; \
		exit 1; \
	fi

release-assets: release-check clean
	@mkdir -p $(DIST_DIR)
	@for target in $(RELEASE_PLATFORMS); do \
		goos=$${target%-*}; \
		goarch=$${target#*-}; \
		archive="mailbox_$${VERSION#v}_$${target}"; \
		pkgdir="$(DIST_DIR)/package"; \
		echo "Building $$archive..."; \
		rm -rf "$$pkgdir"; \
		mkdir -p "$$pkgdir"; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$$goos GOARCH=$$goarch $(GO) build -ldflags="-s -w -X main.version=$(VERSION)" -o "$$pkgdir/$(BINARY_NAME)" . || exit 1; \
		cp README.md "$$pkgdir/"; \
		cp README.zh-CN.md "$$pkgdir/"; \
		tar -czf "$(DIST_DIR)/$$archive.tar.gz" -C "$$pkgdir" . || exit 1; \
		shasum -a 256 "$(DIST_DIR)/$$archive.tar.gz" > "$(DIST_DIR)/$$archive.tar.gz.sha256" || exit 1; \
	done
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz > SHA256SUMS.txt
	@rm -rf $(DIST_DIR)/package
	@echo "Built release artifacts in $(DIST_DIR)/"

release-tag: release-check
	@git rev-parse --verify "$(VERSION)" >/dev/null 2>&1 && { echo "Tag $(VERSION) already exists"; exit 1; } || true
	git tag "$(VERSION)"
	@echo "Created tag $(VERSION)"

release-push: release-check
	git push origin "$(VERSION)"
	@echo "Pushed tag $(VERSION). GitHub Actions will publish the release."

release: check release-tag release-push
	@echo "Release flow started for $(VERSION)"

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
	@echo "  build           - Build binary to bin/ directory"
	@echo "  run             - Build and run the application"
	@echo "  test            - Run tests"
	@echo "  test-v          - Run tests with verbose output"
	@echo "  test-cover      - Run tests with coverage summary"
	@echo "  coverage        - Generate detailed coverage report"
	@echo "  coverage-html   - Generate HTML coverage report"
	@echo "  coverage-pkg    - Show coverage for specific package (PKG=./tui)"
	@echo "  install         - Install to GOPATH/bin"
	@echo "  clean           - Remove build artifacts"
	@echo "  fmt             - Format Go code"
	@echo "  lint            - Run linter"
	@echo "  check           - Run fmt, test, and build"
	@echo "  release-assets  - Build release archives locally (VERSION=v0.1.0)"
	@echo "  release-tag     - Create a git tag locally (VERSION=v0.1.0)"
	@echo "  release-push    - Push an existing release tag (VERSION=v0.1.0)"
	@echo "  release         - Run checks, create tag, and push it (VERSION=v0.1.0)"
	@echo "  dev             - Run with hot reload (requires air)"
	@echo "  help            - Show this help"
