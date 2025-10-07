.PHONY: build install clean test gudcommit gudchangelog

# Build both binaries
build: gudcommit gudchangelog

# Build gudcommit binary
gudcommit:
	@echo "Building gudcommit..."
	@mkdir -p bin
	@go build -o bin/gudcommit ./cmd/gudcommit

# Build gudchangelog binary  
gudchangelog:
	@echo "Building gudchangelog..."
	@mkdir -p bin
	@go build -o bin/gudchangelog ./cmd/gudchangelog

# Install binaries to user's home directory
install: build
	@echo "Installing binaries..."
	@mkdir -p $(HOME)/bin
	@cp bin/gudcommit $(HOME)/bin/
	@cp bin/gudchangelog $(HOME)/bin/
	@echo "✅ Binaries installed to $(HOME)/bin"
	@echo "Add $(HOME)/bin to your PATH if not already there"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o bin/gudcommit-linux-amd64 ./cmd/gudcommit
	@GOOS=linux GOARCH=arm64 go build -o bin/gudcommit-linux-arm64 ./cmd/gudcommit
	@GOOS=darwin GOARCH=amd64 go build -o bin/gudcommit-darwin-amd64 ./cmd/gudcommit
	@GOOS=darwin GOARCH=arm64 go build -o bin/gudcommit-darwin-arm64 ./cmd/gudcommit
	@GOOS=windows GOARCH=amd64 go build -o bin/gudcommit-windows-amd64.exe ./cmd/gudcommit
	@GOOS=windows GOARCH=arm64 go build -o bin/gudcommit-windows-arm64.exe ./cmd/gudcommit
	@GOOS=linux GOARCH=amd64 go build -o bin/gudchangelog-linux-amd64 ./cmd/gudchangelog
	@GOOS=linux GOARCH=arm64 go build -o bin/gudchangelog-linux-arm64 ./cmd/gudchangelog
	@GOOS=darwin GOARCH=amd64 go build -o bin/gudchangelog-darwin-amd64 ./cmd/gudchangelog
	@GOOS=darwin GOARCH=arm64 go build -o bin/gudchangelog-darwin-arm64 ./cmd/gudchangelog
	@GOOS=windows GOARCH=amd64 go build -o bin/gudchangelog-windows-amd64.exe ./cmd/gudchangelog
	@GOOS=windows GOARCH=arm64 go build -o bin/gudchangelog-windows-arm64.exe ./cmd/gudchangelog
	@echo "Cross-platform builds complete!"

# Help target
help:
	@echo "Available targets:"
	@echo "  build       - Build both binaries"
	@echo "  gudcommit   - Build gudcommit binary only"
	@echo "  gudchangelog- Build gudchangelog binary only"
	@echo "  install     - Install binaries to /usr/local/bin"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Download dependencies"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  help        - Show this help message"
