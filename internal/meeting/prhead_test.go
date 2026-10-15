package meeting

import (
	"testing"
)

func TestParsePRNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/owner/repo/pull/12", "12"},
		{"/pull/12", "12"},
		{"pull/12", "12"},
		{"https://github.com/owner/repo/pull/12/files", "12"},
		{"123", "123"},
		{"0", "0"},
		{"", ""},
		{"not-a-pr", ""},
	}

	for _, tt := range tests {
		got := parsePRNumber(tt.input)
		if got != tt.expected {
			t.Errorf("parsePRNumber(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestCheckoutPRHead_MissingPRNumber(t *testing.T) {
	err := checkoutPRHead("", "invalid-pr-url", func(dir string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error on missing PR number, got nil")
	}
}
