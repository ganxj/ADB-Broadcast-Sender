package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config holds application configuration
type Config struct {
	ADBPath       string `json:"adb_path"`
	AutoRefresh   bool   `json:"auto_refresh"`
	RefreshRate   int    `json:"refresh_rate"` // seconds
	LastDeviceID  string `json:"last_device_id"`
	WindowWidth   int    `json:"window_width"`
	WindowHeight  int    `json:"window_height"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ADBPath:       "D:\\Program Files\\Android\\SDK\\platform-tools\\adb.exe",
		AutoRefresh:   true,
		RefreshRate:   5,
		WindowWidth:   900,
		WindowHeight:  700,
	}
}

// Manager manages application configuration
type Manager struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")

	m := &Manager{
		config:     DefaultConfig(),
		configPath: configPath,
	}

	if err := m.Load(); err != nil {
		// If load fails, use defaults and try to save
		_ = m.Save()
	}

	return m, nil
}

// getConfigDir returns the configuration directory path
func getConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", os.ErrNotExist
	}

	configDir := filepath.Join(appData, "adb-broadcast-sender")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// Load loads configuration from file
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.config)
}

// Save saves configuration to file
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Set updates the configuration
func (m *Manager) Set(config *Config) error {
	m.mu.Lock()
	m.config = config
	m.mu.Unlock()
	return m.Save()
}

// SetADBPath sets the ADB path
func (m *Manager) SetADBPath(path string) error {
	m.mu.Lock()
	m.config.ADBPath = path
	m.mu.Unlock()
	return m.Save()
}

// SetLastDevice sets the last used device ID
func (m *Manager) SetLastDevice(deviceID string) error {
	m.mu.Lock()
	m.config.LastDeviceID = deviceID
	m.mu.Unlock()
	return m.Save()
}

// SetWindowSize sets the window dimensions
func (m *Manager) SetWindowSize(width, height int) error {
	m.mu.Lock()
	m.config.WindowWidth = width
	m.config.WindowHeight = height
	m.mu.Unlock()
	return m.Save()
}
