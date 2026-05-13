//go:build property

package models

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestPropertyDeviceDiscoveryAndDisplay validates Property 1 from design document
// Property 1: For any set of connected Android devices, when the application starts or refreshes,
// all devices should be displayed with their unique identifiers and connection states.
// Validates: Requirements 1.1, 1.4, 1.5
func TestPropertyDeviceDiscoveryAndDisplay(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for device serial numbers
	serialGen := gen.RegexMatch(`^[A-Za-z0-9\-_:\.]+$`)

	// Generator for device models
	modelGen := gen.OneConst(
		"Pixel 6",
		"Pixel 7",
		"Galaxy S23",
		"OnePlus 11",
		"Xiaomi 13",
		"",
	)

	// Generator for device states
	stateGen := gen.OneConst("device", "offline", "unauthorized")

	// Generator for connection types
	connectionGen := gen.OneConst("usb", "wifi")

	// Generator for IP addresses
	ipGen := gen.RegexMatch(`^(\d{1,3}\.){3}\d{1,3}$`)

	// Generator for ports
	portGen := gen.IntRange(1024, 65535)

	properties.Property("Device Discovery and Display", prop.ForAll(
		func(serial, model, state, connection, ip string, port int) bool {
			// Create device based on connection type
			var device *Device
			if connection == "wifi" {
				device = NewWiFiDevice(serial, model, state, ip, port)
			} else {
				device = NewDevice(serial, model, state)
			}

			// Validate the device was created correctly
			if device.Serial != serial {
				return false
			}

			if device.Model != model {
				return false
			}

			if device.State != state {
				return false
			}

			if device.Connection != connection {
				return false
			}

			// Check unique identifier (ID should be generated from serial)
			if device.ID == "" {
				return false
			}

			// Check connection state is preserved
			if connection == "wifi" {
				if device.IPAddress != ip || device.Port != port {
					return false
				}
			}

			// Check LastSeen is recent
			if device.LastSeen.IsZero() || device.LastSeen.After(time.Now()) {
				return false
			}

			// Test JSON serialization round-trip
			jsonStr, err := device.ToJSON()
			if err != nil {
				return false
			}

			restored, err := FromJSON(jsonStr)
			if err != nil {
				return false
			}

			// Compare restored device with original
			if restored.Serial != device.Serial {
				return false
			}

			if restored.Model != device.Model {
				return false
			}

			if restored.State != device.State {
				return false
			}

			if restored.Connection != device.Connection {
				return false
			}

			if restored.IPAddress != device.IPAddress {
				return false
			}

			if restored.Port != device.Port {
				return false
			}

			// Test display name generation
			displayName := device.GetDisplayName()
			if displayName == "" {
				return false
			}

			// Test connection info
			connInfo := device.GetConnectionInfo()
			if connInfo == "" {
				return false
			}

			// Test state methods
			switch state {
			case "device":
				if !device.IsConnected() {
					return false
				}
			case "offline":
				if !device.IsOffline() {
					return false
				}
			case "unauthorized":
				if !device.IsUnauthorized() {
					return false
				}
			}

			return true
		},
		serialGen,
		modelGen,
		stateGen,
		connectionGen,
		ipGen,
		portGen,
	))

	properties.TestingRun(t)
}

// TestPropertyDeviceValidation validates device validation logic
func TestPropertyDeviceValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid serial numbers
	validSerialGen := gen.RegexMatch(`^[A-Za-z0-9\-_:\.]{5,20}$`)

	// Generator for invalid serial numbers (empty or too short)
	invalidSerialGen := gen.OneConst("", "a", "ab", "abc", "abcd")

	// Generator for valid states
	validStateGen := gen.OneConst("device", "offline", "unauthorized")

	// Generator for invalid states
	invalidStateGen := gen.OneConst("", "unknown", "disconnected", "ready")

	properties.Property("Device Validation - Valid Devices", prop.ForAll(
		func(serial, model, state string) bool {
			device := NewDevice(serial, model, state)
			err := device.Validate()

			// Valid devices should pass validation
			return err == nil
		},
		validSerialGen,
		gen.AnyString(),
		validStateGen,
	))

	properties.Property("Device Validation - Invalid Serials", prop.ForAll(
		func(serial, model, state string) bool {
			device := NewDevice(serial, model, state)
			err := device.Validate()

			// Devices with invalid serials should fail validation
			return err != nil && err.Error() == ErrInvalidDeviceSerial.Error()
		},
		invalidSerialGen,
		gen.AnyString(),
		validStateGen,
	))

	properties.Property("Device Validation - Invalid States", prop.ForAll(
		func(serial, model, state string) bool {
			device := NewDevice(serial, model, state)
			err := device.Validate()

			// Devices with invalid states should fail validation
			return err != nil && err.Error() == ErrInvalidDeviceState.Error()
		},
		validSerialGen,
		gen.AnyString(),
		invalidStateGen,
	))

	properties.TestingRun(t)
}

// TestPropertyDeviceActiveState validates device active state management
func TestPropertyDeviceActiveState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	serialGen := gen.RegexMatch(`^[A-Za-z0-9\-_:\.]+$`)
	modelGen := gen.AnyString()
	stateGen := gen.OneConst("device", "offline", "unauthorized")

	properties.Property("Device Active State Management", prop.ForAll(
		func(serial, model, state string) bool {
			device := NewDevice(serial, model, state)

			// Initially should not be active
			if device.IsActive {
				return false
			}

			// Set active
			device.SetActive()
			if !device.IsActive {
				return false
			}

			// Set inactive
			device.SetInactive()
			if device.IsActive {
				return false
			}

			// Set active again
			device.SetActive()
			if !device.IsActive {
				return false
			}

			// Update last seen
			oldLastSeen := device.LastSeen
			device.UpdateLastSeen()

			if !device.LastSeen.After(oldLastSeen) {
				return false
			}

			return true
		},
		serialGen,
		modelGen,
		stateGen,
	))

	properties.TestingRun(t)
}

// TestPropertyWiFiDeviceValidation validates WiFi device specific validation
func TestPropertyWiFiDeviceValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	validSerialGen := gen.RegexMatch(`^[A-Za-z0-9\-_:\.]{5,20}$`)
	validIPGen := gen.RegexMatch(`^(\d{1,3}\.){3}\d{1,3}$`)
	validPortGen := gen.IntRange(1024, 65535)
	invalidPortGen := gen.IntRange(0, 1023)

	properties.Property("WiFi Device Validation - Valid WiFi Devices", prop.ForAll(
		func(serial, model, state, ip string, port int) bool {
			device := NewWiFiDevice(serial, model, state, ip, port)
			err := device.Validate()

			// Valid WiFi devices should pass validation
			return err == nil
		},
		validSerialGen,
		gen.AnyString(),
		gen.OneConst("device", "offline", "unauthorized"),
		validIPGen,
		validPortGen,
	))

	properties.Property("WiFi Device Validation - Invalid Ports", prop.ForAll(
		func(serial, model, state, ip string, port int) bool {
			device := NewWiFiDevice(serial, model, state, ip, port)
			err := device.Validate()

			// WiFi devices with invalid ports should fail validation
			return err != nil && err.Error() == ErrInvalidWiFiConnection.Error()
		},
		validSerialGen,
		gen.AnyString(),
		gen.OneConst("device", "offline", "unauthorized"),
		validIPGen,
		invalidPortGen,
	))

	properties.TestingRun(t)
}
