package sitlog

import (
	"testing"
)

func TestRedactEmpty(t *testing.T) {
	if got := Redact(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestRedactNoSecrets(t *testing.T) {
	in := "plain log message with normal words and numbers 12345"
	if got := Redact(in); got != in {
		t.Fatalf("expected %q, got %q", in, got)
	}
}

func TestRedactEnvKeys(t *testing.T) {
	keys := []string{
		"JULES_API_KEY",
		"XAI_API_KEY",
		"GOOGLE_API_KEY",
		"GEMINI_API_KEY",
	}

	for _, k := range keys {
		t.Run(k, func(t *testing.T) {
			secret := "mock-secret-" + k
			t.Setenv(k, secret)
			in := "auth header with " + secret + " embedded"
			want := "auth header with [REDACTED] embedded"
			if got := Redact(in); got != want {
				t.Fatalf("Redact(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

func TestRedactTokenShapes(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"ghp", "ghp_12345678901234567890"},
		{"gho", "gho_12345678901234567890"},
		{"ghu", "ghu_12345678901234567890"},
		{"ghs", "ghs_12345678901234567890"},
		{"ghr", "ghr_12345678901234567890"},
		{"github_pat", "github_pat_12345678901234567890_extra"},
		{"xai", "xai-12345678901234567890"},
		{"AIza", "AIzaSyD123456789012345678901234567"},
		{"Bearer mixed case", "bearer abcdefghijklmnopqrstuvwxyz"},
		{"BEARER uppercase", "BEARER abcdefghijklmnopqrstuvwxyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := "prefix " + tt.token + " suffix"
			want := "prefix [REDACTED] suffix"
			if got := Redact(in); got != want {
				t.Fatalf("Redact(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

func TestRedactTokenShapesNegative(t *testing.T) {
	// Tokens below length thresholds or non-matching prefixes should not be redacted
	tests := []struct {
		name string
		text string
	}{
		{"ghp too short", "ghp_short123"},
		{"github_pat too short", "github_pat_short"},
		{"xai too short", "xai-short"},
		{"AIza too short", "AIzaShort12345"},
		{"Bearer too short", "Bearer short123"},
		{"unrelated prefix", "ghx_12345678901234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Redact(tt.text); got != tt.text {
				t.Fatalf("Redact(%q) = %q, want %q", tt.text, got, tt.text)
			}
		})
	}
}

func TestRedactMultipleSecrets(t *testing.T) {
	t.Setenv("JULES_API_KEY", "jules-sec-123")
	t.Setenv("XAI_API_KEY", "xai-sec-456")

	in := "keys: jules-sec-123 and xai-sec-456 and token ghp_12345678901234567890"
	want := "keys: [REDACTED] and [REDACTED] and token [REDACTED]"
	if got := Redact(in); got != want {
		t.Fatalf("Redact(%q) = %q, want %q", in, got, want)
	}
}
