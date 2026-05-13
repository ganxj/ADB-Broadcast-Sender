package app

import (
	"context"
	"sync"
	"time"

	"adb-broadcast-sender/internal/adb"
	"adb-broadcast-sender/internal/config"
	"adb-broadcast-sender/internal/models"
)

// AppState holds the application state
type AppState struct {
	ConfigManager    *config.Manager
	DeviceManager    *adb.DeviceManager
	BroadcastManager *adb.BroadcastManager

	SelectedDevice *models.Device
	IsRefreshing   bool

	mu              sync.RWMutex
	onStateChange   func()
}

// NewAppState creates a new application state
func NewAppState() (*AppState, error) {
	// Initialize config manager
	configManager, err := config.NewManager()
	if err != nil {
		return nil, err
	}

	// Initialize device manager
	deviceManager, err := adb.NewDeviceManager(configManager.Get().ADBPath)
	if err != nil {
		return nil, err
	}

	// Initialize broadcast manager
	broadcastManager, err := adb.NewBroadcastManager()
	if err != nil {
		return nil, err
	}

	state := &AppState{
		ConfigManager:    configManager,
		DeviceManager:    deviceManager,
		BroadcastManager: broadcastManager,
	}

	// Set device change callback
	deviceManager.SetDeviceChangeCallback(func() {
		state.notifyStateChange()
	})

	return state, nil
}

// SetStateChangeCallback sets the callback for state changes
func (s *AppState) SetStateChangeCallback(callback func()) {
	s.mu.Lock()
	s.onStateChange = callback
	s.mu.Unlock()
}

// notifyStateChange notifies listeners of state changes
func (s *AppState) notifyStateChange() {
	s.mu.RLock()
	callback := s.onStateChange
	s.mu.RUnlock()

	if callback != nil {
		callback()
	}
}

// RefreshDevices refreshes the device list
func (s *AppState) RefreshDevices() error {
	s.mu.Lock()
	s.IsRefreshing = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.IsRefreshing = false
		s.mu.Unlock()
		s.notifyStateChange()
	}()

	return s.DeviceManager.RefreshDevices()
}

// SelectDevice selects a device
func (s *AppState) SelectDevice(device *models.Device) {
	s.mu.Lock()
	s.SelectedDevice = device
	s.mu.Unlock()

	if device != nil {
		s.ConfigManager.SetLastDevice(device.Serial)
	}

	s.notifyStateChange()
}

// SelectDeviceBySerial selects a device by serial
func (s *AppState) SelectDeviceBySerial(serial string) {
	device := s.DeviceManager.GetDevice(serial)
	s.SelectDevice(device)
}

// SendBroadcast sends a broadcast to the selected device
func (s *AppState) SendBroadcast(content string) (*models.Broadcast, error) {
	s.mu.RLock()
	device := s.SelectedDevice
	s.mu.RUnlock()

	if device == nil {
		return nil, ErrNoDeviceSelected
	}

	result, err := s.BroadcastManager.SendBroadcast(device.Serial, content)
	s.notifyStateChange()
	return result, err
}

// SendBroadcastWithAction sends a broadcast with custom action
func (s *AppState) SendBroadcastWithAction(action, content string) (*models.Broadcast, error) {
	s.mu.RLock()
	device := s.SelectedDevice
	s.mu.RUnlock()

	if device == nil {
		return nil, ErrNoDeviceSelected
	}

	result, err := s.BroadcastManager.SendBroadcastWithAction(device.Serial, action, content)
	s.notifyStateChange()
	return result, err
}

// GetDevices returns the device list
func (s *AppState) GetDevices() []*models.Device {
	return s.DeviceManager.GetDevices()
}

// GetConnectedDevices returns only connected devices
func (s *AppState) GetConnectedDevices() []*models.Device {
	return s.DeviceManager.GetConnectedDevices()
}

// GetBroadcastHistory returns the broadcast history
func (s *AppState) GetBroadcastHistory() []*models.Broadcast {
	return s.BroadcastManager.GetHistory()
}

// ClearHistory clears the broadcast history
func (s *AppState) ClearHistory() {
	s.BroadcastManager.ClearHistory()
	s.notifyStateChange()
}

// StartAutoRefresh starts automatic device refresh
func (s *AppState) StartAutoRefresh(ctx context.Context) {
	cfg := s.ConfigManager.Get()
	if cfg.AutoRefresh && cfg.RefreshRate > 0 {
		s.DeviceManager.StartWatching(ctx, time.Duration(cfg.RefreshRate)*time.Second)
	}
}

// GetSelectedDevice returns the selected device
func (s *AppState) GetSelectedDevice() *models.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SelectedDevice
}

// IsDeviceRefreshing returns true if devices are being refreshed
func (s *AppState) IsDeviceRefreshing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsRefreshing
}
