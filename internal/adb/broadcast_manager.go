package adb

import (
	"fmt"
	"strings"
	"sync"

	"adb-broadcast-sender/internal/models"

	"github.com/electricbubble/gadb"
)

// BroadcastManager manages broadcast sending
type BroadcastManager struct {
	client     gadb.Client
	history    []*models.Broadcast
	mu         sync.RWMutex
	maxHistory int
}

// NewBroadcastManager creates a new broadcast manager
func NewBroadcastManager() (*BroadcastManager, error) {
	client, err := gadb.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create ADB client: %w", err)
	}

	return &BroadcastManager{
		client:     client,
		history:    make([]*models.Broadcast, 0),
		maxHistory: 100,
	}, nil
}

// getDeviceBySerial finds a device by its serial number
func (bm *BroadcastManager) getDeviceBySerial(serial string) (gadb.Device, error) {
	deviceList, err := bm.client.DeviceList()
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

// SendBroadcast sends a broadcast to a device
func (bm *BroadcastManager) SendBroadcast(deviceSerial, content string) (*models.Broadcast, error) {
	broadcast := models.NewBroadcast(content, deviceSerial)

	// Validate content
	if err := broadcast.ValidateContent(); err != nil {
		broadcast.MarkFailed(err.Error())
		bm.addToHistory(broadcast)
		return broadcast, err
	}

	// Get device
	device, err := bm.getDeviceBySerial(deviceSerial)
	if err != nil {
		broadcast.MarkFailed(fmt.Sprintf("Device not found: %s", deviceSerial))
		bm.addToHistory(broadcast)
		return broadcast, fmt.Errorf("device not found: %s", deviceSerial)
	}

	// Build and execute command
	escapedContent := strings.ReplaceAll(content, `"`, `\"`)
	shellCmd := fmt.Sprintf(`am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "%s"`, escapedContent)

	output, err := device.RunShellCommand(shellCmd)
	if err != nil {
		broadcast.MarkFailed(err.Error())
		bm.addToHistory(broadcast)
		return broadcast, err
	}

	// Parse output to determine success
	if strings.Contains(output, "result=0") || strings.Contains(output, "Broadcast completed") {
		broadcast.MarkSuccess(output)
	} else {
		broadcast.MarkFailed(output)
	}

	bm.addToHistory(broadcast)
	return broadcast, nil
}

// SendBroadcastWithAction sends a broadcast with custom action
func (bm *BroadcastManager) SendBroadcastWithAction(deviceSerial, action, content string) (*models.Broadcast, error) {
	broadcast := models.NewBroadcast(content, deviceSerial)

	// Validate content
	if err := broadcast.ValidateContent(); err != nil {
		broadcast.MarkFailed(err.Error())
		bm.addToHistory(broadcast)
		return broadcast, err
	}

	// Get device
	device, err := bm.getDeviceBySerial(deviceSerial)
	if err != nil {
		broadcast.MarkFailed(fmt.Sprintf("Device not found: %s", deviceSerial))
		bm.addToHistory(broadcast)
		return broadcast, fmt.Errorf("device not found: %s", deviceSerial)
	}

	// Build custom command
	escapedContent := strings.ReplaceAll(content, `"`, `\"`)
	shellCmd := fmt.Sprintf(`am broadcast -a %s --es data "%s"`, action, escapedContent)

	output, err := device.RunShellCommand(shellCmd)
	if err != nil {
		broadcast.MarkFailed(err.Error())
		bm.addToHistory(broadcast)
		return broadcast, err
	}

	// Parse output
	if strings.Contains(output, "result=0") || strings.Contains(output, "Broadcast completed") {
		broadcast.MarkSuccess(output)
	} else {
		broadcast.MarkFailed(output)
	}

	bm.addToHistory(broadcast)
	return broadcast, nil
}

// SendRawBroadcast sends a raw broadcast command
func (bm *BroadcastManager) SendRawBroadcast(deviceSerial, command string) (*models.Broadcast, error) {
	broadcast := models.NewBroadcast(command, deviceSerial)

	// Get device
	device, err := bm.getDeviceBySerial(deviceSerial)
	if err != nil {
		broadcast.MarkFailed(fmt.Sprintf("Device not found: %s", deviceSerial))
		bm.addToHistory(broadcast)
		return broadcast, fmt.Errorf("device not found: %s", deviceSerial)
	}

	output, err := device.RunShellCommand(command)
	if err != nil {
		broadcast.MarkFailed(err.Error())
		bm.addToHistory(broadcast)
		return broadcast, err
	}

	// Parse output
	if strings.Contains(output, "result=0") || strings.Contains(output, "Broadcast completed") {
		broadcast.MarkSuccess(output)
	} else {
		broadcast.MarkFailed(output)
	}

	bm.addToHistory(broadcast)
	return broadcast, nil
}

// GetHistory returns the broadcast history
func (bm *BroadcastManager) GetHistory() []*models.Broadcast {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	history := make([]*models.Broadcast, len(bm.history))
	copy(history, bm.history)
	return history
}

// ClearHistory clears the broadcast history
func (bm *BroadcastManager) ClearHistory() {
	bm.mu.Lock()
	bm.history = make([]*models.Broadcast, 0)
	bm.mu.Unlock()
}

// addToHistory adds a broadcast to history
func (bm *BroadcastManager) addToHistory(broadcast *models.Broadcast) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.history = append(bm.history, broadcast)

	// Limit history size
	if len(bm.history) > bm.maxHistory {
		bm.history = bm.history[len(bm.history)-bm.maxHistory:]
	}
}

// SetMaxHistory sets the maximum history size
func (bm *BroadcastManager) SetMaxHistory(max int) {
	bm.mu.Lock()
	bm.maxHistory = max
	bm.mu.Unlock()
}
