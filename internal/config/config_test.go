package config

import (
	"reflect"
	"testing"
)

func TestStripForbidden(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "all three forbidden keys plus PATH",
			input: []string{
				"XAI_API_KEY=xai_value",
				"PATH=/usr/bin",
				"GOOGLE_API_KEY=google_value",
				"GEMINI_API_KEY=gemini_value",
			},
			expected: []string{
				"PATH=/usr/bin",
			},
		},
		{
			name: "forbidden keys without equals sign",
			input: []string{
				"XAI_API_KEY",
				"PATH=/usr/bin",
			},
			expected: []string{
				"PATH=/usr/bin",
			},
		},
		{
			name: "partial match is kept",
			input: []string{
				"XAI_API_KEY_BACKUP=1",
				"PATH=/usr/bin",
			},
			expected: []string{
				"XAI_API_KEY_BACKUP=1",
				"PATH=/usr/bin",
			},
		},
		{
			name: "no equals sign and not forbidden",
			input: []string{
				"JUST_A_STRING",
			},
			expected: []string{
				"JUST_A_STRING",
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the input to ensure the original is not modified
			originalInput := make([]string, len(tt.input))
			copy(originalInput, tt.input)

			result := StripForbidden(tt.input)

			// Check if the original input was modified
			if len(tt.input) > 0 && !reflect.DeepEqual(tt.input, originalInput) {
				t.Errorf("StripForbidden modified the input slice. Original: %v, Current: %v", originalInput, tt.input)
			}

			// Check if the result is correct
			if !reflect.DeepEqual(result, tt.expected) {
				// Handle nil slice returned instead of empty slice
				if len(result) == 0 && len(tt.expected) == 0 {
					return
				}
				t.Errorf("StripForbidden() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStripForbiddenOriginalSliceNotModified(t *testing.T) {
	input := []string{"XAI_API_KEY=123", "VALID=abc"}

	// Create a backing array to explicitly check if elements are overwritten
	original := make([]string, 2)
	copy(original, input)

	_ = StripForbidden(input)

	if input[0] != original[0] || input[1] != original[1] {
		t.Errorf("Original slice was modified. Input changed from %v to %v", original, input)
	}
}
