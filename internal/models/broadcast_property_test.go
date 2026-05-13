//go:build property

package models

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestPropertyBroadcastCommandConstruction validates Property 3 from design document
// Property 3: For any valid broadcast content string, when constructing the ADB command,
// the system should produce the exact format: adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "content"
// Validates: Requirements 2.2
func TestPropertyBroadcastCommandConstruction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for broadcast content (excluding dangerous characters)
	contentGen := gen.SuchThat(
		func(s string) bool {
			// Exclude strings with backticks or command substitution
			return !strings.Contains(s, "`") && !strings.Contains(s, "$(")
		},
		gen.AnyString(),
	)

	// Generator for device serials
	serialGen := gen.RegexMatch(`^[A-Za-z0-9\-_:\.]+$`)

	properties.Property("Broadcast Command Construction - Basic Format", prop.ForAll(
		func(content string) bool {
			broadcast := NewBroadcast(content, "device_123")
			command := broadcast.BuildCommand()

			// Check command starts with the correct prefix
			expectedPrefix := `adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "`
			if !strings.HasPrefix(command, expectedPrefix) {
				return false
			}

			// Check command ends with a quote
			if !strings.HasSuffix(command, `"`) {
				return false
			}

			// Extract the content from the command
			contentStart := len(expectedPrefix)
			contentEnd := len(command) - 1 // Remove trailing quote
			extractedContent := command[contentStart:contentEnd]

			// The extracted content should be the original content with quotes escaped
			expectedContent := strings.ReplaceAll(content, `"`, `\"`)
			if extractedContent != expectedContent {
				return false
			}

			// Test that the command can be reconstructed
			reconstructed := `adb shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "` + expectedContent + `"`
			if command != reconstructed {
				return false
			}

			return true
		},
		contentGen,
	))

	properties.Property("Broadcast Command Construction - With Device Serial", prop.ForAll(
		func(content, serial string) bool {
			broadcast := NewBroadcast(content, "device_"+serial)
			command := broadcast.BuildCommandWithDevice(serial)

			// Check command starts with the correct prefix including device serial
			expectedPrefix := `adb -s ` + serial + ` shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "`
			if !strings.HasPrefix(command, expectedPrefix) {
				return false
			}

			// Check command ends with a quote
			if !strings.HasSuffix(command, `"`) {
				return false
			}

			// Extract the content from the command
			contentStart := len(expectedPrefix)
			contentEnd := len(command) - 1 // Remove trailing quote
			extractedContent := command[contentStart:contentEnd]

			// The extracted content should be the original content with quotes escaped
			expectedContent := strings.ReplaceAll(content, `"`, `\"`)
			if extractedContent != expectedContent {
				return false
			}

			// Test that the command can be reconstructed
			reconstructed := `adb -s ` + serial + ` shell am broadcast -a com.sunmi.scanner.ACTION_DATA_CODE_RECEIVED --es data "` + expectedContent + `"`
			if command != reconstructed {
				return false
			}

			return true
		},
		contentGen,
		serialGen,
	))

	properties.Property("Broadcast Command Construction - Quote Escaping", prop.ForAll(
		func(content string) bool {
			// Only test strings that contain quotes
			if !strings.Contains(content, `"`) {
				return true // Skip if no quotes
			}

			broadcast := NewBroadcast(content, "device_123")
			command := broadcast.BuildCommand()

			// Count quotes in original content
			originalQuoteCount := strings.Count(content, `"`)

			// Count quotes in command (excluding the surrounding quotes)
			commandWithoutSurrounding := command[strings.Index(command, `"`)+1 : len(command)-1]
			escapedQuoteCount := strings.Count(commandWithoutSurrounding, `\"`)

			// The number of escaped quotes should match the number of original quotes
			if escapedQuoteCount != originalQuoteCount {
				return false
			}

			// The command should not contain unescaped quotes (except the surrounding ones)
			unescapedQuoteCount := strings.Count(commandWithoutSurrounding, `"`)
			if unescapedQuoteCount > 0 {
				return false
			}

			return true
		},
		gen.AnyString(), // Allow any string including those with quotes
	))

	properties.TestingRun(t)
}

// TestPropertyBroadcastContentValidation validates broadcast content validation
func TestPropertyBroadcastContentValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid content (non-empty, no dangerous chars)
	validContentGen := gen.SuchThat(
		func(s string) bool {
			return strings.TrimSpace(s) != "" &&
				!strings.Contains(s, "`") &&
				!strings.Contains(s, "$(") &&
				len(s) <= 1024*1024 // 1MB limit
		},
		gen.AnyString(),
	)

	// Generator for empty or whitespace content
	emptyContentGen := gen.OneConst("", " ", "  ", "\t", "\n", "\r\n")

	// Generator for content with dangerous characters
	dangerousContentGen := gen.SuchThat(
		func(s string) bool {
			return strings.Contains(s, "`") || strings.Contains(s, "$(")
		},
		gen.AnyString(),
	)

	properties.Property("Broadcast Content Validation - Valid Content", prop.ForAll(
		func(content string) bool {
			broadcast := NewBroadcast(content, "device_123")
			err := broadcast.ValidateContent()

			// Valid content should pass validation
			return err == nil
		},
		validContentGen,
	))

	properties.Property("Broadcast Content Validation - Empty Content", prop.ForAll(
		func(content string) bool {
			broadcast := NewBroadcast(content, "device_123")
			err := broadcast.ValidateContent()

			// Empty content should fail validation with specific error
			return err != nil && err.Error() == ErrEmptyBroadcastContent.Error()
		},
		emptyContentGen,
	))

	properties.Property("Broadcast Content Validation - Dangerous Content", prop.ForAll(
		func(content string) bool {
			broadcast := NewBroadcast(content, "device_123")
			err := broadcast.ValidateContent()

			// Content with dangerous characters should fail validation
			return err != nil && err.Error() == ErrInvalidBroadcastContent.Error()
		},
		dangerousContentGen,
	))

	properties.TestingRun(t)
}

// TestPropertyBroadcastStateManagement validates broadcast state management
func TestPropertyBroadcastStateManagement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	contentGen := gen.AnyString()
	deviceIDGen := gen.RegexMatch(`^device_[A-Za-z0-9\-_:\.]+$`)

	properties.Property("Broadcast State Management - Success/Failure States", prop.ForAll(
		func(content, deviceID string) bool {
			broadcast := NewBroadcast(content, deviceID)

			// Initially should be pending
			if !broadcast.IsPending() {
				return false
			}

			if broadcast.IsSuccess() {
				return false
			}

			if broadcast.IsFailed() {
				return false
			}

			// Mark as success
			output := "Test output"
			broadcast.MarkSuccess(output)

			if !broadcast.IsSuccess() {
				return false
			}

			if broadcast.IsPending() {
				return false
			}

			if broadcast.IsFailed() {
				return false
			}

			if broadcast.Output != output {
				return false
			}

			if broadcast.Error != "" {
				return false
			}

			// Mark as failed
			errorMsg := "Test error"
			broadcast.MarkFailed(errorMsg)

			if !broadcast.IsFailed() {
				return false
			}

			if broadcast.IsSuccess() {
				return false
			}

			if broadcast.IsPending() {
				return false
			}

			if broadcast.Error != errorMsg {
				return false
			}

			return true
		},
		contentGen,
		deviceIDGen,
	))

	properties.TestingRun(t)
}

// TestPropertyBroadcastJSONRoundTrip validates JSON serialization/deserialization
func TestPropertyBroadcastJSONRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	contentGen := gen.AnyString()
	deviceIDGen := gen.RegexMatch(`^device_[A-Za-z0-9\-_:\.]+$`)
	resultGen := gen.OneConst("success", "failed", "pending")
	outputGen := gen.AnyString()
	errorGen := gen.AnyString()

	properties.Property("Broadcast JSON Round-Trip", prop.ForAll(
		func(content, deviceID, result, output, errorMsg string) bool {
			// Create broadcast with all fields
			broadcast := NewBroadcastWithResult(content, deviceID, result, output, errorMsg)

			// Serialize to JSON
			jsonStr, err := broadcast.ToJSON()
			if err != nil {
				return false
			}

			// Deserialize from JSON
			restored, err := BroadcastFromJSON(jsonStr)
			if err != nil {
				return false
			}

			// Compare fields
			if restored.Content != broadcast.Content {
				return false
			}

			if restored.DeviceID != broadcast.DeviceID {
				return false
			}

			if restored.Result != broadcast.Result {
				return false
			}

			if restored.Output != broadcast.Output {
				return false
			}

			if restored.Error != broadcast.Error {
				return false
			}

			// Check that IDs match
			if restored.ID != broadcast.ID {
				return false
			}

			// Check timestamps are close (within 1 second)
			timeDiff := broadcast.Timestamp.Sub(restored.Timestamp)
			if timeDiff > 1e9 || timeDiff < -1e9 { // 1 second in nanoseconds
				return false
			}

			return true
		},
		contentGen,
		deviceIDGen,
		resultGen,
		outputGen,
		errorGen,
	))

	properties.TestingRun(t)
}
