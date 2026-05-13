package models

import (
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	device := NewDevice("1234567890", "Pixel 6", "device")

	if device.Serial != "1234567890" {
		t.Errorf("Expected serial '1234567890', got '%s'", device.Serial)
	}

	if device.Model != "Pixel 6" {
		t.Errorf("Expected model 'Pixel 6', got '%s'", device.Model)
	}

	if device.State != "device" {
		t.Errorf("Expected state 'device', got '%s'", device.State)
	}

	if device.Connection != "usb" {
		t.Errorf("Expected connection 'usb', got '%s'", device.Connection)
	}

	if !device.LastSeen.After(time.Now().Add(-1 * time.Second)) {
		t.Error("LastSeen should be recent")
	}

	if device.IsActive {
		t.Error("New device should not be active by default")
	}
}

func TestNewWiFiDevice(t *testing.T) {
	device := NewWiFiDevice("1234567890", "Pixel 6", "device", "192.168.1.100", 5555)

	if device.Connection != "wifi" {
		t.Errorf("Expected connection 'wifi', got '%s'", device.Connection)
	}

	if device.IPAddress != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", device.IPAddress)
	}

	if device.Port != 5555 {
		t.Errorf("Expected port 5555, got %d", device.Port)
	}
}

func TestDeviceJSONSerialization(t *testing.T) {
	original := NewDevice("1234567890", "Pixel 6", "device")
	original.IsActive = true

	jsonStr, err := original.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	restored, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}

	if restored.Serial != original.Serial {
		t.Errorf("Serial mismatch: expected '%s', got '%s'", original.Serial, restored.Serial)
	}

	if restored.Model != original.Model {
		t.Errorf("Model mismatch: expected '%s', got '%s'", original.Model, restored.Model)
	}

	if restored.State != original.State {
		t.Errorf("State mismatch: expected '%s', got '%s'", original.State, restored.State)
	}

	if restored.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch: expected %v, got %v", original.IsActive, restored.IsActive)
	}
}

func TestDeviceStateMethods(t *testing.T) {
	tests := []struct {
		state          string
		isConnected    bool
		isOffline      bool
		isUnauthorized bool
	}{
		{"device", true, false, false},
		{"offline", false, true, false},
		{"unauthorized", false, false, true},
		{"unknown", false, false, false},
	}

	for _, tt := range tests {
		device := NewDevice("1234567890", "Pixel 6", tt.state)

		if device.IsConnected() != tt.isConnected {
			t.Errorf("IsConnected() for state '%s': expected %v, got %v", tt.state, tt.isConnected, device.IsConnected())
		}

		if device.IsOffline() != tt.isOffline {
			t.Errorf("IsOffline() for state '%s': expected %v, got %v", tt.state, tt.isOffline, device.IsOffline())
		}

		if device.IsUnauthorized() != tt.isUnauthorized {
			t.Errorf("IsUnauthorized() for state '%s': expected %v, got %v", tt.state, tt.isUnauthorized, device.IsUnauthorized())
		}
	}
}

func TestDeviceActiveState(t *testing.T) {
	device := NewDevice("1234567890", "Pixel 6", "device")

	if device.IsActive {
		t.Error("Device should not be active initially")
	}

	device.SetActive()
	if !device.IsActive {
		t.Error("Device should be active after SetActive()")
	}

	device.SetInactive()
	if device.IsActive {
		t.Error("Device should not be active after SetInactive()")
	}
}

func TestDeviceDisplayName(t *testing.T) {
	device := NewDevice("1234567890", "Pixel 6", "device")
	expected := "Pixel 6 (1234567890)"

	if device.GetDisplayName() != expected {
		t.Errorf("Expected display name '%s', got '%s'", expected, device.GetDisplayName())
	}

	device2 := NewDevice("1234567890", "", "device")
	expected2 := "1234567890"

	if device2.GetDisplayName() != expected2 {
		t.Errorf("Expected display name '%s', got '%s'", expected2, device2.GetDisplayName())
	}
}

func TestDeviceConnectionInfo(t *testing.T) {
	usbDevice := NewDevice("1234567890", "Pixel 6", "device")
	if usbDevice.GetConnectionInfo() != "USB" {
		t.Errorf("Expected 'USB', got '%s'", usbDevice.GetConnectionInfo())
	}

	wifiDevice := NewWiFiDevice("1234567890", "Pixel 6", "device", "192.168.1.100", 5555)
	expected := "WiFi: 192.168.1.100:5555"
	if wifiDevice.GetConnectionInfo() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, wifiDevice.GetConnectionInfo())
	}
}

func TestDeviceValidation(t *testing.T) {
	tests := []struct {
		name        string
		device      *Device
		expectError bool
	}{
		{
			name:        "Valid USB device",
			device:      NewDevice("1234567890", "Pixel 6", "device"),
			expectError: false,
		},
		{
			name:        "Valid WiFi device",
			device:      NewWiFiDevice("1234567890", "Pixel 6", "device", "192.168.1.100", 5555),
			expectError: false,
		},
		{
			name:        "Invalid serial",
			device:      &Device{Serial: "", Model: "Pixel 6", State: "device", Connection: "usb"},
			expectError: true,
		},
		{
			name:        "Invalid state",
			device:      &Device{Serial: "1234567890", Model: "Pixel 6", State: "", Connection: "usb"},
			expectError: true,
		},
		{
			name:        "Invalid connection type",
			device:      &Device{Serial: "1234567890", Model: "Pixel 6", State: "device", Connection: "bluetooth"},
			expectError: true,
		},
		{
			name:        "Invalid WiFi connection - missing IP",
			device:      &Device{Serial: "1234567890", Model: "Pixel 6", State: "device", Connection: "wifi", Port: 5555},
			expectError: true,
		},
		{
			name:        "Invalid WiFi connection - invalid port",
			device:      &Device{Serial: "1234567890", Model: "Pixel 6", State: "device", Connection: "wifi", IPAddress: "192.168.1.100", Port: 0},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
