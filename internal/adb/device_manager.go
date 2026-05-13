package adb

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"adb-broadcast-sender/internal/models"

	"github.com/electricbubble/gadb"
)

// DeviceManager manages connected Android devices
type DeviceManager struct {
	client          gadb.Client
	adbPath         string
	devices         map[string]*models.Device
	mu              sync.RWMutex
	onDeviceChange  func()
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(adbPath string) (*DeviceManager, error) {
	client, err := gadb.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create ADB client: %w", err)
	}

	return &DeviceManager{
		client:  client,
		adbPath: adbPath,
		devices: make(map[string]*models.Device),
	}, nil
}

// SetDeviceChangeCallback sets the callback for device changes
func (dm *DeviceManager) SetDeviceChangeCallback(callback func()) {
	dm.mu.Lock()
	dm.onDeviceChange = callback
	dm.mu.Unlock()
}

// RefreshDevices refreshes the device list
func (dm *DeviceManager) RefreshDevices() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Get device list from ADB
	deviceList, err := dm.client.DeviceList()
	if err != nil {
		return fmt.Errorf("failed to get device list: %w", err)
	}

	// Clear old devices
	dm.devices = make(map[string]*models.Device)

	// Process each device
	for _, d := range deviceList {
		serial := d.Serial()
		state, stateErr := d.State()

		device := &models.Device{
			Serial:    serial,
			State:     string(state),
			LastSeen:  time.Now(),
		}

		// Determine connection type
		if strings.Contains(serial, ":") {
			device.Connection = "wifi"
			parts := strings.Split(serial, ":")
			if len(parts) == 2 {
				device.IPAddress = parts[0]
				fmt.Sscanf(parts[1], "%d", &device.Port)
			}
		} else {
			device.Connection = "usb"
		}

		// Get device model
		if model, err := d.Model(); err == nil {
			device.Model = model
		}

		// Update state if there was an error
		if stateErr != nil {
			device.State = "unknown"
		}

		device.ID = "device_" + serial
		dm.devices[device.Serial] = device
	}

	// Notify callback
	if dm.onDeviceChange != nil {
		go dm.onDeviceChange()
	}

	return nil
}

// ConnectWiFi connects to a device via WiFi
func (dm *DeviceManager) ConnectWiFi(ip string, port int) error {
	err := dm.client.Connect(ip, port)
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d: %w", ip, port, err)
	}

	// Refresh device list after connecting
	return dm.RefreshDevices()
}

// DisconnectWiFi disconnects from a WiFi device
func (dm *DeviceManager) DisconnectWiFi(ip string, port int) error {
	err := dm.client.Disconnect(ip, port)
	if err != nil {
		return fmt.Errorf("failed to disconnect from %s:%d: %w", ip, port, err)
	}

	// Refresh device list after disconnecting
	return dm.RefreshDevices()
}

// GetDevices returns all connected devices
func (dm *DeviceManager) GetDevices() []*models.Device {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	devices := make([]*models.Device, 0, len(dm.devices))
	for _, d := range dm.devices {
		devices = append(devices, d)
	}
	return devices
}

// GetDevice returns a device by serial
func (dm *DeviceManager) GetDevice(serial string) *models.Device {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.devices[serial]
}

// GetConnectedDevices returns only connected devices (state = "device")
func (dm *DeviceManager) GetConnectedDevices() []*models.Device {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	devices := make([]*models.Device, 0)
	for _, d := range dm.devices {
		if d.IsConnected() {
			devices = append(devices, d)
		}
	}
	return devices
}

// GetDeviceClient returns the gadb.Device for a given serial
func (dm *DeviceManager) GetDeviceClient(serial string) (gadb.Device, error) {
	deviceList, err := dm.client.DeviceList()
	if err != nil {
		return gadb.Device{}, fmt.Errorf("failed to get device list: %w", err)
	}

	for _, d := range deviceList {
		if d.Serial() == serial {
			return d, nil
		}
	}

	return gadb.Device{}, fmt.Errorf("device not found: %s", serial)
}

// StartWatching starts watching for device changes
func (dm *DeviceManager) StartWatching(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				dm.RefreshDevices()
			}
		}
	}()
}
