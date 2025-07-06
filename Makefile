# OpenTask Makefile
# Build and development automation for OpenTask CLI

# Variables
BINARY_NAME=opentask
VERSION?=0.1.0
BUILD_DIR=build
DIST_DIR=dist
GO_VERSION=1.24
LDFLAGS=-ldflags="-X 'main.Version=${VERSION}'"

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	@echo "Building ${BINARY_NAME} v${VERSION}..."
	@mkdir -p ${BUILD_DIR}
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} .

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p ${DIST_DIR}
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-arm64 .

# Build for macOS
.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p ${DIST_DIR}
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-darwin-arm64 .

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p ${DIST_DIR}
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-windows-arm64.exe .

# Install locally
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME}..."
	cp ${BUILD_DIR}/${BINARY_NAME} /usr/local/bin/

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf ${BUILD_DIR}
	rm -rf ${DIST_DIR}
	rm -f coverage.out coverage.html

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the application
.PHONY: run
run:
	@echo "Running ${BINARY_NAME}..."
	go run . $(ARGS)

# Watch for changes and rebuild
.PHONY: watch
watch:
	@echo "Watching for changes..."
	which air > /dev/null || go install github.com/air-verse/air@latest
	air

# Create release
.PHONY: release
release: clean build-all
	@echo "Creating release v${VERSION}..."
	@mkdir -p ${DIST_DIR}/release
	@for binary in ${DIST_DIR}/${BINARY_NAME}-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip -j ${DIST_DIR}/release/$$(basename $$binary .exe).zip $$binary; \
		else \
			tar -czf ${DIST_DIR}/release/$$(basename $$binary).tar.gz -C ${DIST_DIR} $$(basename $$binary); \
		fi; \
	done
	@echo "Release artifacts created in ${DIST_DIR}/release/"

# Help
.PHONY: help
help:
	@echo "OpenTask Build System"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build for current platform"
	@echo "  build-all     Build for all platforms"
	@echo "  build-linux   Build for Linux (amd64, arm64)"
	@echo "  build-darwin  Build for macOS (amd64, arm64)"
	@echo "  build-windows Build for Windows (amd64, arm64)"
	@echo "  install       Install locally to /usr/local/bin"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  lint          Run linter"
	@echo "  fmt           Format code"
	@echo "  tidy          Tidy dependencies"
	@echo "  clean         Clean build artifacts"
	@echo "  dev-setup     Setup development environment"
	@echo "  run           Run the application (use ARGS=... for arguments)"
	@echo "  watch         Watch for changes and rebuild"
	@echo "  release       Create release artifacts"
	@echo "  help          Show this help"