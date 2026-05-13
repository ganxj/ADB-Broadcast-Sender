package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Broadcast represents a broadcast intent sent to an Android device
type Broadcast struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Result    string    `json:"result"` // "success", "failed", "pending"
	Output    string    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// NewBroadcast creates a new Broadcast instance
func NewBroadcast(content, deviceID string) *Broadcast {
	return &Broadcast{
		ID:        generateBroadcastID(),
		Content:   content,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Result:    "pending",
	}
}

// NewBroadcastWithResult creates a new Broadcast instance with result
func NewBroadcastWithResult(content, deviceID, result, output, errorMsg string) *Broadcast {
	broadcast := NewBroadcast(content, deviceID)
	broadcast.Result = result
	broadcast.Output = output
	broadcast.Error = errorMsg
	return broadcast
}

// BuildCommand constructs the ADB broadcast command
// Format: adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "content"
func (b *Broadcast) BuildCommand() string {
	// Escape quotes in content for shell command
	escapedContent := strings.ReplaceAll(b.Content, `"`, `\"`)

	// Build the command according to the specified format
	return fmt.Sprintf(`adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "%s"`, escapedContent)
}

// BuildCommandWithDevice constructs the ADB broadcast command with device serial
// Format: adb -s <serial> shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "content"
func (b *Broadcast) BuildCommandWithDevice(deviceSerial string) string {
	// Escape quotes in content for shell command
	escapedContent := strings.ReplaceAll(b.Content, `"`, `\"`)

	// Build the command with device serial
	return fmt.Sprintf(`adb -s %s shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "%s"`, deviceSerial, escapedContent)
}

// ValidateContent validates the broadcast content
func (b *Broadcast) ValidateContent() error {
	if strings.TrimSpace(b.Content) == "" {
		return ErrEmptyBroadcastContent
	}

	// Check for maximum length (Android intent extra size limit)
	if len(b.Content) > 1024*1024 { // 1MB limit
		return ErrBroadcastContentTooLarge
	}

	// Check for potentially dangerous characters
	if strings.Contains(b.Content, "`") || strings.Contains(b.Content, "$(") {
		return ErrInvalidBroadcastContent
	}

	return nil
}

// MarkSuccess marks the broadcast as successful
func (b *Broadcast) MarkSuccess(output string) {
	b.Result = "success"
	b.Output = output
	b.Error = ""
}

// MarkFailed marks the broadcast as failed
func (b *Broadcast) MarkFailed(errorMsg string) {
	b.Result = "failed"
	b.Error = errorMsg
}

// IsSuccess returns true if broadcast was successful
func (b *Broadcast) IsSuccess() bool {
	return b.Result == "success"
}

// IsFailed returns true if broadcast failed
func (b *Broadcast) IsFailed() bool {
	return b.Result == "failed"
}

// IsPending returns true if broadcast is pending
func (b *Broadcast) IsPending() bool {
	return b.Result == "pending"
}

// GetStatus returns a human-readable status
func (b *Broadcast) GetStatus() string {
	switch b.Result {
	case "success":
		return "Success"
	case "failed":
		return "Failed"
	case "pending":
		return "Pending"
	default:
		return "Unknown"
	}
}

// GetFormattedTimestamp returns formatted timestamp
func (b *Broadcast) GetFormattedTimestamp() string {
	return b.Timestamp.Format("2006-01-02 15:04:05")
}

// ToJSON converts Broadcast to JSON string
func (b *Broadcast) ToJSON() (string, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON creates Broadcast from JSON string
func BroadcastFromJSON(jsonStr string) (*Broadcast, error) {
	var broadcast Broadcast
	err := json.Unmarshal([]byte(jsonStr), &broadcast)
	if err != nil {
		return nil, err
	}
	return &broadcast, nil
}

// Validate checks if the broadcast has valid data
func (b *Broadcast) Validate() error {
	if err := b.ValidateContent(); err != nil {
		return err
	}

	if b.DeviceID == "" {
		return ErrInvalidDeviceID
	}

	if b.Result != "success" && b.Result != "failed" && b.Result != "pending" {
		return ErrInvalidBroadcastResult
	}

	return nil
}

// GetSummary returns a summary of the broadcast
func (b *Broadcast) GetSummary() string {
	return fmt.Sprintf("Broadcast to device %s: %s (%s)", b.DeviceID, b.GetStatus(), b.GetFormattedTimestamp())
}

// Helper function to generate broadcast ID
func generateBroadcastID() string {
	// Use timestamp-based ID
	return fmt.Sprintf("broadcast_%d", time.Now().UnixNano())
}

// Error definitions
var (
	ErrEmptyBroadcastContent    = errors.New("broadcast content cannot be empty")
	ErrBroadcastContentTooLarge = errors.New("broadcast content is too large")
	ErrInvalidBroadcastContent  = errors.New("broadcast content contains invalid characters")
	ErrInvalidDeviceID          = errors.New("invalid device ID")
	ErrInvalidBroadcastResult   = errors.New("invalid broadcast result")
)
