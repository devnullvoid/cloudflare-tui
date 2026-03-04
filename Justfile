# Justfile for cftui

# Default recipe to display help
default:
    @just --list

# Run tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Run linter
lint:
    golangci-lint run

# Build binary
build:
    go build -o cftui ./cmd/cftui

# Install binary to $GOPATH/bin
install:
    go install ./cmd/cftui

# Clean build artifacts
clean:
    rm -f cftui

# Format code
fmt:
    go fmt ./...

# Run all checks (fmt, lint, test)
check: fmt lint test

# Build for multiple platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -o dist/cftui-linux-amd64 ./cmd/cftui
    GOOS=linux GOARCH=arm64 go build -o dist/cftui-linux-arm64 ./cmd/cftui
    GOOS=darwin GOARCH=amd64 go build -o dist/cftui-darwin-amd64 ./cmd/cftui
    GOOS=darwin GOARCH=arm64 go build -o dist/cftui-darwin-arm64 ./cmd/cftui
    GOOS=windows GOARCH=amd64 go build -o dist/cftui-windows-amd64.exe ./cmd/cftui

# Record demo with VHS
demo:
    cd assets && vhs demo.tape
