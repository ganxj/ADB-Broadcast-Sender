# ADB Broadcast Sender

A Windows application written in Go for sending broadcast intents to Android devices via ADB. The application provides a web-based interface accessible through your browser.

## Features

- **ADB Connection Management**: Show connected devices, establish WiFi connections
- **Broadcast Sending**: Send broadcast intents to selected devices
- **Configuration**: Customize ADB path and application settings
- **Web Interface**: Modern web-based interface accessible via browser
- **Real-time Updates**: Server-Sent Events for live device status updates

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

# Build the application (disables CGO to avoid build issues)
make build

# Run the application
make run

# Run tests
make test
```

### Using Go commands directly:
```bash
# Build for Windows (disable CGO to avoid build issues)
set CGO_ENABLED=0
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
│   ├── adb-broadcast-sender/
│   │   └── main.go          # Web application entry point (HTTP server)
│   └── test-cli/
│       └── main.go          # CLI tool for ADB connectivity testing
├── internal/
│   ├── models/              # Data models (Device, Broadcast) with tests
│   ├── config/              # Configuration management
│   ├── adb/                 # ADB integration layer
│   └── app/                 # Business logic layer
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── Makefile                 # Build automation
├── README.md                # This file
└── verify_models.md         # Model validation status
```

## Usage

1. **Launch the application**:
   ```bash
   # Build and run
   make build && ./adb-broadcast-sender.exe
   
   # Or run directly
   go run ./cmd/adb-broadcast-sender
   ```

2. **Access the web interface**:
   - Open your browser and go to: `http://localhost:18088`
   - The application will automatically open the browser for you

3. **Connect Android device**:
   - Connect device via USB or WiFi
   - Click "Refresh" to see connected devices

4. **Send broadcast**:
   - Select a device from the device list
   - Enter broadcast content
   - Click "Send Broadcast" to send to selected device

### Troubleshooting

**Port 8080 already in use**:
If you see "bind: Only one usage of each socket address" error, port 8080 is already in use. You can:
1. Stop the other application using port 8080
2. Or modify the port in `cmd/adb-broadcast-sender/main.go` line 84

**ADB not found**:
Ensure ADB is installed and in your PATH. Default path: `D:\Program Files\Android\SDK\platform-tools\adb.exe`

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