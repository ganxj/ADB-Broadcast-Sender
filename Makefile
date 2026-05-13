# ADB Broadcast Sender Makefile

.PHONY: build run test clean

# Build the application for Windows (disable CGO to avoid build issues)
build:
	set CGO_ENABLED=0 && go build -o adb-broadcast-sender.exe ./cmd/adb-broadcast-sender

# Run the application
run:
	go run ./cmd/adb-broadcast-sender

# Run tests
test:
	go test ./...

# Run property-based tests
test-property:
	go test -tags=property ./...

# Clean build artifacts
clean:
	rm -f adb-broadcast-sender.exe
	rm -rf dist/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for Windows with release settings
release:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/adb-broadcast-sender.exe ./cmd/adb-broadcast-sender

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run all tests"
	@echo "  test-property - Run property-based tests"
	@echo "  deps         - Install dependencies"
	@echo "  release      - Build optimized release version"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"