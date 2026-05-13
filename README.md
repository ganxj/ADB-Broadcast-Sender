# ADB Broadcast Sender

A Windows desktop application written in Go for sending broadcast intents to Android devices via ADB.

## Features

- **ADB Connection Management**: Show connected devices, establish WiFi connections
- **Broadcast Sending**: Send broadcast intents to selected devices
- **Configuration**: Customize ADB path and application settings
- **Modern GUI**: Clean, intuitive Windows interface using Fyne framework

## Prerequisites

1. **Go 1.21 or later**: [Download and install Go](https://go.dev/dl/)
2. **ADB (Android Debug Bridge)**: Ensure ADB is installed and in your PATH
   - Default path: `D:\Program Files\Android\SDK\platform-tools\adb.exe`
3. **Windows**: The application is designed for Windows

## Installation

1. Install Go from [https://go.dev/dl/](https://go.dev/dl/)
2. Clone or download this project
3. Open terminal in project directory
4. Install dependencies:
   ```bash
   go mod download
   ```

## Building and Running

### Using Makefile (recommended):
```bash
# Install dependencies
make deps

# Build the application
make build

# Run the application
make run

# Run tests
make test
```

### Using Go commands directly:
```bash
# Build for Windows
go build -o adb-broadcast-sender.exe ./cmd/adb-broadcast-sender

# Run the application
go run ./cmd/adb-broadcast-sender

# Run tests
go test ./...
```

## Project Structure

```
adb-broadcast-sender/
├── cmd/
│   └── adb-broadcast-sender/
│       └── main.go          # Application entry point
├── internal/
│   ├── models/              # Data models (Device, Broadcast)
│   ├── config/              # Configuration management
│   ├── adb/                 # ADB integration layer
│   └── app/                 # Business logic layer
├── pkg/                     # Public packages (if needed)
├── config/                  # Configuration files
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── Makefile                 # Build automation
└── README.md                # This file
```

## Usage

1. Launch the application
2. Connect Android device via USB or WiFi
3. Select device from the device list
4. Enter broadcast content
5. Click "Send Broadcast" to send to selected device

## Configuration

The application stores configuration in:
- Windows: `%APPDATA%\adb-broadcast-sender\config.json`

You can configure:
- ADB executable path
- Default connection settings
- UI preferences
- History settings

## Development

### Adding Dependencies
```bash
go get <package-name>
```

### Running Tests
```bash
# Unit tests
go test ./...

# Property-based tests
go test -tags=property ./...
```

### Building for Release
```bash
make release
```

## License

[Add your license here]

## Support

For issues and feature requests, please create an issue in the project repository.