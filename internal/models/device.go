package models

import (
	"encoding/json"
	"errors"
	"time"
)

// Device represents an Android device connected via ADB
type Device struct {
	ID         string    `json:"id"`
	Serial     string    `json:"serial"`
	Model      string    `json:"model"`
	State      string    `json:"state"`      // "device", "offline", "unauthorized"
	Connection string    `json:"connection"` // "usb", "wifi"
	IPAddress  string    `json:"ip_address,omitempty"`
	Port       int       `json:"port,omitempty"`
	LastSeen   time.Time `json:"last_seen"`
	IsActive   bool      `json:"is_active"`
}

// NewDevice creates a new Device instance
func NewDevice(serial, model, state string) *Device {
	return &Device{
		ID:         generateDeviceID(serial),
		Serial:     serial,
		Model:      model,
		State:      state,
		Connection: "usb", // default to USB
		LastSeen:   time.Now(),
		IsActive:   false,
	}
}

// NewWiFiDevice creates a new WiFi-connected Device instance
func NewWiFiDevice(serial, model, state, ipAddress string, port int) *Device {
	device := NewDevice(serial, model, state)
	device.Connection = "wifi"
	device.IPAddress = ipAddress
	device.Port = port
	return device
}

// ToJSON converts Device to JSON string
func (d *Device) ToJSON() (string, error) {
	bytes, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON creates Device from JSON string
func FromJSON(jsonStr string) (*Device, error) {
	var device Device
	err := json.Unmarshal([]byte(jsonStr), &device)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// IsConnected returns true if device is in "device" or "online" state
func (d *Device) IsConnected() bool {
	return d.State == "device" || d.State == "online"
}

// IsOffline returns true if device is in "offline" state
func (d *Device) IsOffline() bool {
	return d.State == "offline"
}

// IsUnauthorized returns true if device is in "unauthorized" state
func (d *Device) IsUnauthorized() bool {
	return d.State == "unauthorized"
}

// UpdateLastSeen updates the LastSeen timestamp to current time
func (d *Device) UpdateLastSeen() {
	d.LastSeen = time.Now()
}

// SetActive marks this device as active
func (d *Device) SetActive() {
	d.IsActive = true
}

// SetInactive marks this device as inactive
func (d *Device) SetInactive() {
	d.IsActive = false
}

// GetDisplayName returns a display-friendly name for the device
func (d *Device) GetDisplayName() string {
	if d.Model != "" {
		return d.Model + " (" + d.Serial + ")"
	}
	return d.Serial
}

// GetConnectionInfo returns connection information string
func (d *Device) GetConnectionInfo() string {
	if d.Connection == "wifi" {
		return "WiFi: " + d.IPAddress + ":" + string(rune(d.Port))
	}
	return "USB"
}

// Validate checks if the device has valid data
func (d *Device) Validate() error {
	if d.Serial == "" {
		return ErrInvalidDeviceSerial
	}
	if d.State == "" {
		return ErrInvalidDeviceState
	}
	if d.Connection != "usb" && d.Connection != "wifi" {
		return ErrInvalidConnectionType
	}
	if d.Connection == "wifi" && (d.IPAddress == "" || d.Port <= 0) {
		return ErrInvalidWiFiConnection
	}
	return nil
}

// Helper function to generate device ID
func generateDeviceID(serial string) string {
	// Simple hash-based ID generation
	// In production, use a proper UUID or hash
	return "device_" + serial
}

// Error definitions
var (
	ErrInvalidDeviceSerial   = errors.New("invalid device serial")
	ErrInvalidDeviceState    = errors.New("invalid device state")
	ErrInvalidConnectionType = errors.New("invalid connection type")
	ErrInvalidWiFiConnection = errors.New("invalid WiFi connection details")
)
