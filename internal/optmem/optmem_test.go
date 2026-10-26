package optmem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createFakeMemo(t *testing.T, script string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	logFile := filepath.Join(dir, "run.log")
	envFile := filepath.Join(dir, "env.log")
	binPath := filepath.Join(dir, "grok-agy-memo")

	content := "#!/bin/sh\n" +
		"env > " + envFile + "\n" +
		"echo \"$@\" >> " + logFile + "\n" +
		script
	if err := os.WriteFile(binPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir, logFile
}

func TestMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	out, warn := Wake(t.TempDir())
	if out != "" || warn == "" {
		t.Fatalf("expected empty out and non-empty warn, got %q, %q", out, warn)
	}
	if warn := Note(t.TempDir(), "hello"); warn == "" {
		t.Fatal("expected non-empty warn for missing binary")
	}
}

func TestWakeAndNote(t *testing.T) {
	t.Setenv("GROK_AGY", "1")
	t.Setenv("GROK_AGY_JOB_DIR", "/tmp/job")
	t.Setenv("MEMORY_DIR", "/tmp/mem")
	t.Setenv("GEMINI_API_KEY", "secret-gemini")

	script := `
echo "warning: ignored stderr" >&2
if [ "$1" = "wake" ]; then
    echo "line 1"
    echo "line 2"
    exit 0
fi
if [ "$1" = "note" ]; then
    exit 0
fi
`
	dir, logFile := createFakeMemo(t, script)

	out, warn := Wake(dir)
	if warn != "" {
		t.Fatalf("unexpected warning on exit 0: %q", warn)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || lines[0] != "line 1" || lines[1] != "line 2" {
		t.Fatalf("expected 2 lines, got %q", out)
	}

	if warn := Note(dir, "test note"); warn != "" {
		t.Fatalf("unexpected warning on note exit 0: %q", warn)
	}

	logged, err := os.ReadFile(logFile)
	if err != nil || !strings.Contains(string(logged), "note test note") {
		t.Fatalf("note args not recorded: %v, logged: %s", err, string(logged))
	}

	envBytes, err := os.ReadFile(filepath.Join(dir, "env.log"))
	if err != nil {
		t.Fatal(err)
	}
	envLines := strings.Split(string(envBytes), "\n")
	for _, forbidden := range []string{"GROK_AGY=", "GROK_AGY_JOB_DIR=", "MEMORY_DIR=", "GEMINI_API_KEY="} {
		for _, line := range envLines {
			if strings.HasPrefix(line, forbidden) {
				t.Fatalf("child env contains forbidden %q", forbidden)
			}
		}
	}
}

func TestNoteOnce(t *testing.T) {
	resetNoted()
	dir := t.TempDir()
	flagFile := filepath.Join(dir, "flag.txt")
	counterFile := filepath.Join(dir, "count.txt")
	script := "if [ ! -f " + flagFile + " ]; then\n  touch " + flagFile + "\n  exit 1\nfi\necho 1 >> " + counterFile + "\nexit 0\n"
	memoDir, _ := createFakeMemo(t, script)

	if w0 := NoteOnce(memoDir, "fail"); w0 == "" {
		t.Fatalf("expected warning on failure")
	}
	w1 := NoteOnce(memoDir, "first")
	if w1 != "" {
		t.Fatalf("unexpected warning: %q", w1)
	}
	w2 := NoteOnce(memoDir, "second")
	if w2 != "" {
		t.Fatalf("unexpected warning: %q", w2)
	}

	data, err := os.ReadFile(counterFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(lines))
	}
}
