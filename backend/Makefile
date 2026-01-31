# Backend Makefile

.PHONY: build test lint clean run

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Run the server locally
run:
	go run ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate go.sum
tidy:
	go mod tidy
