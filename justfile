# Noteleaf project commands

# Default recipe - show available commands
default:
    @just --list

# Run all tests
test:
    go test ./...

# Run tests with coverage
coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run tests and show coverage in terminal
cov:
    go test ./... -coverprofile=coverage.out
    go tool cover -func=coverage.out

# Build the binary to /tmp/
build:
    mkdir -p ./tmp/
    go build -o ./tmp/noteleaf ./cmd/
    @echo "Binary built: ./tmp/noteleaf"

# Clean build artifacts
clean:
    rm -f coverage.out coverage.html
    rm -rf /tmp/noteleaf

# Run linting
lint:
    go vet ./...
    go fmt ./...

# Run all quality checks
check: lint cov

# Install dependencies
deps:
    go mod download
    go mod tidy

# Run the application (after building)
run: build
    /tmp/noteleaf/noteleaf

# Show project status
status:
    @echo "Go version:"
    @go version
    @echo ""
    @echo "Module info:"
    @go list -m
    @echo ""
    @echo "Dependencies:"
    @go list -m all | head -10

# Quick development workflow
dev: clean lint test build
    @echo "Development workflow complete!"
