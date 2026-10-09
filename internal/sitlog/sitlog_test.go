package sitlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenCreatesTimestampLog(t *testing.T) {
	cwd := t.TempDir()
	now := time.Date(2026, 10, 9, 12, 30, 5, 0, time.UTC)
	l, err := Open(cwd, now)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	want := filepath.Join(cwd, "logs", "council-20261009-123005.log")
	if l.Path != want {
		t.Fatalf("path %s want %s", l.Path, want)
	}
	b, err := os.ReadFile(l.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "sitting log started") {
		t.Fatalf("start line missing: %q", b)
	}
}

func TestRotateKeepsFifteenMatchingOnly(t *testing.T) {
	cwd := t.TempDir()
	dir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 15; i++ {
		name := filepath.Join(dir, time.Date(2020, 1, i, 0, 0, 0, 0, time.UTC).Format("council-20060102-150405.log"))
		if err := os.WriteFile(name, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	notes := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(notes, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := Open(cwd, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
	matches, err := filepath.Glob(filepath.Join(dir, "council-*.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 15 {
		t.Fatalf("matching %d want 15: %v", len(matches), matches)
	}
	if _, err := os.Stat(notes); err != nil {
		t.Fatalf("notes.txt removed: %v", err)
	}
	oldest := filepath.Join(dir, "council-20200101-000000.log")
	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Fatalf("oldest still present")
	}
}

func TestRedactEnvAndToken(t *testing.T) {
	t.Setenv("JULES_API_KEY", "secret-jules-value")
	in := "key=secret-jules-value token=ghp_abcdefghijklmnopqrstuvwxyz123456"
	out := Redact(in)
	if strings.Contains(out, "secret-jules-value") {
		t.Fatalf("env not redacted: %q", out)
	}
	if strings.Contains(out, "ghp_") {
		t.Fatalf("token not redacted: %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("missing marker: %q", out)
	}
}

func TestNilLogLineNoPanic(t *testing.T) {
	var l *Log
	l.Line("sys", "x")
	l.Close()
}
