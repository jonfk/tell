# Justfile for TELL (Terminal English Language Liaison)

# Default recipe to run when just is called without arguments
default:
    @just --list

# Build the project
build:
    go build -o bin/tell ./cmd/tell

# Run tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Install the binary to $GOPATH/bin
install:
    go install ./cmd/tell

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    go vet ./...
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run; \
    else \
        echo "golangci-lint not installed, skipping additional linting"; \
    fi

# Generate shell integration scripts
generate-shell-scripts:
    mkdir -p scripts
    go run ./cmd/tell/main.go --generate-shell-scripts

# Run with specific arguments
run *ARGS:
    go run ./cmd/tell/main.go {{ARGS}}

# Update dependencies
update-deps:
    go get -u ./...
    go mod tidy
