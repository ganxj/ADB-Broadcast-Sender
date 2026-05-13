package models

import (
	"strings"
	"testing"
	"time"
)

func TestNewBroadcast(t *testing.T) {
	content := "test broadcast content"
	deviceID := "device_1234567890"

	broadcast := NewBroadcast(content, deviceID)

	if broadcast.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, broadcast.Content)
	}

	if broadcast.DeviceID != deviceID {
		t.Errorf("Expected device ID '%s', got '%s'", deviceID, broadcast.DeviceID)
	}

	if broadcast.Result != "pending" {
		t.Errorf("Expected result 'pending', got '%s'", broadcast.Result)
	}

	if !broadcast.Timestamp.After(time.Now().Add(-1 * time.Second)) {
		t.Error("Timestamp should be recent")
	}

	if broadcast.ID == "" {
		t.Error("Broadcast ID should not be empty")
	}
}

func TestNewBroadcastWithResult(t *testing.T) {
	content := "test broadcast content"
	deviceID := "device_1234567890"
	output := "Broadcast completed successfully"
	errorMsg := ""

	broadcast := NewBroadcastWithResult(content, deviceID, "success", output, errorMsg)

	if broadcast.Result != "success" {
		t.Errorf("Expected result 'success', got '%s'", broadcast.Result)
	}

	if broadcast.Output != output {
		t.Errorf("Expected output '%s', got '%s'", output, broadcast.Output)
	}

	if broadcast.Error != errorMsg {
		t.Errorf("Expected error '%s', got '%s'", errorMsg, broadcast.Error)
	}
}

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Simple content",
			content:  "test123",
			expected: `adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "test123"`,
		},
		{
			name:     "Content with spaces",
			content:  "test broadcast content",
			expected: `adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "test broadcast content"`,
		},
		{
			name:     "Content with quotes",
			content:  `test "quoted" content`,
			expected: `adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "test \"quoted\" content"`,
		},
		{
			name:     "Empty content",
			content:  "",
			expected: `adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcast := NewBroadcast(tt.content, "device_123")
			command := broadcast.BuildCommand()

			if command != tt.expected {
				t.Errorf("Expected command:\n%s\nGot:\n%s", tt.expected, command)
			}
		})
	}
}

func TestBuildCommandWithDevice(t *testing.T) {
	content := "test content"
	deviceSerial := "1234567890"

	broadcast := NewBroadcast(content, "device_123")
	command := broadcast.BuildCommandWithDevice(deviceSerial)

	expected := `adb -s 1234567890 shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "test content"`

	if command != expected {
		t.Errorf("Expected command:\n%s\nGot:\n%s", expected, command)
	}
}

func TestValidateContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "Valid content",
			content:     "test broadcast",
			expectError: false,
		},
		{
			name:        "Empty content",
			content:     "",
			expectError: true,
		},
		{
			name:        "Whitespace only",
			content:     "   ",
			expectError: true,
		},
		{
			name:        "Content with backticks",
			content:     "test `command`",
			expectError: true,
		},
		{
			name:        "Content with command substitution",
			content:     "test $(command)",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcast := NewBroadcast(tt.content, "device_123")
			err := broadcast.ValidateContent()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestBroadcastStatusMethods(t *testing.T) {
	tests := []struct {
		result     string
		isSuccess  bool
		isFailed   bool
		isPending  bool
		statusText string
	}{
		{"success", true, false, false, "Success"},
		{"failed", false, true, false, "Failed"},
		{"pending", false, false, true, "Pending"},
		{"unknown", false, false, false, "Unknown"},
	}

	for _, tt := range tests {
		broadcast := NewBroadcastWithResult("content", "device_123", tt.result, "", "")

		if broadcast.IsSuccess() != tt.isSuccess {
			t.Errorf("IsSuccess() for result '%s': expected %v, got %v", tt.result, tt.isSuccess, broadcast.IsSuccess())
		}

		if broadcast.IsFailed() != tt.isFailed {
			t.Errorf("IsFailed() for result '%s': expected %v, got %v", tt.result, tt.isFailed, broadcast.IsFailed())
		}

		if broadcast.IsPending() != tt.isPending {
			t.Errorf("IsPending() for result '%s': expected %v, got %v", tt.result, tt.isPending, broadcast.IsPending())
		}

		if broadcast.GetStatus() != tt.statusText {
			t.Errorf("GetStatus() for result '%s': expected '%s', got '%s'", tt.result, tt.statusText, broadcast.GetStatus())
		}
	}
}

func TestMarkSuccessAndFailed(t *testing.T) {
	broadcast := NewBroadcast("content", "device_123")

	// Initially should be pending
	if !broadcast.IsPending() {
		t.Error("New broadcast should be pending")
	}

	// Mark as success
	output := "Broadcast completed successfully"
	broadcast.MarkSuccess(output)

	if !broadcast.IsSuccess() {
		t.Error("Broadcast should be marked as success")
	}

	if broadcast.Output != output {
		t.Errorf("Expected output '%s', got '%s'", output, broadcast.Output)
	}

	if broadcast.Error != "" {
		t.Errorf("Expected empty error, got '%s'", broadcast.Error)
	}

	// Mark as failed
	errorMsg := "ADB command failed"
	broadcast.MarkFailed(errorMsg)

	if !broadcast.IsFailed() {
		t.Error("Broadcast should be marked as failed")
	}

	if broadcast.Error != errorMsg {
		t.Errorf("Expected error '%s', got '%s'", errorMsg, broadcast.Error)
	}
}

func TestBroadcastJSONSerialization(t *testing.T) {
	original := NewBroadcastWithResult("test content", "device_1234567890", "success", "output message", "")

	jsonStr, err := original.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	restored, err := BroadcastFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}

	if restored.Content != original.Content {
		t.Errorf("Content mismatch: expected '%s', got '%s'", original.Content, restored.Content)
	}

	if restored.DeviceID != original.DeviceID {
		t.Errorf("DeviceID mismatch: expected '%s', got '%s'", original.DeviceID, restored.DeviceID)
	}

	if restored.Result != original.Result {
		t.Errorf("Result mismatch: expected '%s', got '%s'", original.Result, restored.Result)
	}

	if restored.Output != original.Output {
		t.Errorf("Output mismatch: expected '%s', got '%s'", original.Output, restored.Output)
	}

	// Check timestamps are close (within 1 second)
	timeDiff := original.Timestamp.Sub(restored.Timestamp)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("Timestamp mismatch too large: %v", timeDiff)
	}
}

func TestBroadcastValidation(t *testing.T) {
	tests := []struct {
		name        string
		broadcast   *Broadcast
		expectError bool
	}{
		{
			name:        "Valid broadcast",
			broadcast:   NewBroadcast("valid content", "device_123"),
			expectError: false,
		},
		{
			name:        "Empty content",
			broadcast:   NewBroadcast("", "device_123"),
			expectError: true,
		},
		{
			name:        "Invalid device ID",
			broadcast:   NewBroadcast("content", ""),
			expectError: true,
		},
		{
			name: "Invalid result",
			broadcast: &Broadcast{
				Content:  "content",
				DeviceID: "device_123",
				Result:   "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.broadcast.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetSummary(t *testing.T) {
	broadcast := NewBroadcast("test content", "device_1234567890")
	broadcast.MarkSuccess("output")

	summary := broadcast.GetSummary()

	if !strings.Contains(summary, "Broadcast to device device_1234567890") {
		t.Errorf("Summary should contain device ID, got: %s", summary)
	}

	if !strings.Contains(summary, "Success") {
		t.Errorf("Summary should contain status, got: %s", summary)
	}

	if !strings.Contains(summary, broadcast.GetFormattedTimestamp()) {
		t.Errorf("Summary should contain timestamp, got: %s", summary)
	}
}

func TestGetFormattedTimestamp(t *testing.T) {
	// Create a broadcast with a specific timestamp
	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	broadcast := &Broadcast{
		Timestamp: testTime,
	}

	formatted := broadcast.GetFormattedTimestamp()
	expected := "2023-12-25 10:30:45"

	if formatted != expected {
		t.Errorf("Expected formatted timestamp '%s', got '%s'", expected, formatted)
	}
}
