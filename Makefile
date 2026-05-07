.PHONY: build test lint clean install deps help

# Default target
help:
	@echo "Spur - Environment Reproduction Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  build    - Build the spur binary"
	@echo "  test     - Run all tests"
	@echo "  lint     - Run linter (requires golangci-lint)"
	@echo "  clean    - Remove build artifacts"
	@echo "  install  - Install spur to GOPATH/bin"
	@echo "  deps     - Install dependencies"

# Build the binary
build:
	@echo "Building spur..."
	go build -o spur .

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v -race -coverprofile=coverage.out

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f spur
	rm -f *.yaml
	rm -f coverage.out

# Install to GOPATH/bin
install:
	@echo "Installing spur..."
	go install .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
