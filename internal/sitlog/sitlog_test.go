package sitlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenCreatesLog(t *testing.T) {
	cwd := t.TempDir()
	now := time.Now()
	l, err := Open(cwd, now)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	logsDir := filepath.Join(cwd, "logs")
	entries, _ := os.ReadDir(logsDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "council-") {
		t.Fatalf("unexpected name %s", entries[0].Name())
	}
}

func TestRotate(t *testing.T) {
	cwd := t.TempDir()
	logsDir := filepath.Join(cwd, "logs")
	os.MkdirAll(logsDir, 0755)

	for i := 0; i < 15; i++ {
		name := fmt.Sprintf("council-20230101-%06d.log", i)
		os.WriteFile(filepath.Join(logsDir, name), []byte("test"), 0644)
	}
	os.WriteFile(filepath.Join(logsDir, "notes.txt"), []byte("notes"), 0644)

	l, err := Open(cwd, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	l.Close()

	entries, _ := os.ReadDir(logsDir)
	var logs []string
	var notes int
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "council-") && strings.HasSuffix(e.Name(), ".log") {
			logs = append(logs, e.Name())
		} else if e.Name() == "notes.txt" {
			notes++
		}
	}

	if len(logs) != 15 {
		t.Fatalf("expected 15 logs, got %d", len(logs))
	}
	if notes != 1 {
		t.Fatalf("expected notes.txt to remain")
	}
	// the oldest (000000) should be gone
	if logs[0] == "council-20230101-000000.log" {
		t.Fatalf("oldest log was not deleted")
	}
}

func TestRedact(t *testing.T) {
	os.Setenv("JULES_API_KEY", "supersecret123")
	defer os.Unsetenv("JULES_API_KEY")

	tests := []struct {
		in, out string
	}{
		{"my key is supersecret123!", "my key is [REDACTED]!"},
		{"ghp_12345678901234567890xx", "[REDACTED]"},
		{"github_pat_11AAAAAAA01234567890_abc", "[REDACTED]"},
		{"xai-12345678901234567890xx", "[REDACTED]"},
		{"AIzaSy123456789012345678901234567890", "[REDACTED]"},
		{"Authorization: Bearer 12345678901234567890xx", "Authorization: Bearer [REDACTED]"},
		{"short AIza123", "short AIza123"}, // not long enough
	}

	for _, tt := range tests {
		got := Redact(tt.in)
		if got != tt.out {
			t.Errorf("Redact(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

func TestNilLog(t *testing.T) {
	var l *Log
	l.Line("sys", "hi") // should not panic
	l.Close()
}

func TestRedactShortThenLong(t *testing.T) {
	in := "Found ghp_ prefix before the actual token: ghp_12345678901234567890xx"
	out := "Found ghp_ prefix before the actual token: [REDACTED]"
	got := Redact(in)
	if got != out {
		t.Errorf("Redact(%q) = %q, want %q", in, got, out)
	}
}
