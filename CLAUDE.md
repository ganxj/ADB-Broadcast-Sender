# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Windows desktop application written in Go for sending broadcast intents to Android devices via ADB. The application uses the Fyne framework for its GUI and the gadb library for ADB integration.

## Common Commands

### Build and Run
```bash
# Build the application
make build
# or
go build -o adb-broadcast-sender.exe ./cmd/adb-broadcast-sender

# Run the application
make run
# or
go run ./cmd/adb-broadcast-sender

# Run the CLI test (verifies ADB connectivity)
go run ./cmd/test-cli
```

### Testing
```bash
# Run all unit tests
make test
# or
go test ./...

# Run property-based tests (requires build tag)
make test-property
# or
go test -tags=property ./...
```

### Development
```bash
# Install dependencies
make deps
# or
go mod download && go mod tidy

# Build optimized release version
make release

# Clean build artifacts
make clean
```

## Architecture

### Layered Structure
1. **Models** (`internal/models/`): Core data structures (Device, Broadcast) with validation logic
2. **Configuration** (`internal/config/`): Application settings and persistence
3. **ADB Integration** (`internal/adb/`): Device discovery and broadcast sending via ADB
4. **Application Logic** (`internal/app/`): Business logic and state management
5. **GUI** (`cmd/adb-broadcast-sender/`): Fyne-based Windows interface

### Key Dependencies
- `fyne.io/fyne/v2`: GUI framework for Windows desktop applications
- `github.com/electricbubble/gadb`: ADB client library for Go
- `github.com/leanovate/gopter`: Property-based testing framework

## Development Notes

### ADB Configuration
- Default ADB path: `D:\Program Files\Android\SDK\platform-tools\adb.exe`
- Configuration stored in: `%APPDATA%\adb-broadcast-sender\config.json`
- The CLI test (`cmd/test-cli/`) verifies ADB connectivity before GUI work

### Testing Strategy
- Unit tests for all model methods
- Property-based tests for validation logic (use `-tags=property`)
- Models are currently the only completed component (see `verify_models.md`)

### Current Implementation Status
- ✅ Core models (Device, Broadcast) with tests
- ⏳ GUI skeleton (basic Fyne window)
- ⏳ Configuration system (placeholder)
- ⏳ ADB integration layer (placeholder)
- ⏳ Application logic layer (placeholder)

### Windows-Specific Considerations
- Application is Windows-only (Fyne GUI)
- Build with `GOOS=windows GOARCH=amd64` for releases
- Uses CGO for OpenGL rendering (Fyne dependency)

## Workflow

1. **Model Changes**: Update `internal/models/`, run `go test ./...` and `go test -tags=property ./...`
2. **ADB Integration**: Use `cmd/test-cli` to verify ADB connectivity before implementing GUI features
3. **GUI Development**: Test with `make run` to see Fyne interface changes
4. **Configuration**: Settings are stored in Windows AppData directory

## Important Paths
- Main application: `cmd/adb-broadcast-sender/main.go`
- CLI test: `cmd/test-cli/main.go`
- Model definitions: `internal/models/device.go`, `internal/models/broadcast.go`
- Build configuration: `Makefile`