package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestAgyRunnerArgvDangerouslySkipPermissions(t *testing.T) {
	r := &AgyRunner{CLI: "agy", Model: "m", Cwd: t.TempDir()}
	argv := r.argv(worker.Job{Text: "task", Paths: nil})

	foundFlag := false
	for i, a := range argv {
		if a == "--dangerously-skip-permissions" {
			foundFlag = true
			if i >= len(argv)-2 {
				t.Fatalf("--dangerously-skip-permissions must be before -p")
			}
		}
	}
	if !foundFlag {
		t.Fatalf("argv missing --dangerously-skip-permissions: %v", argv)
	}
}

func TestAgyRunnerDeniedActionsInReport(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	fakeJSON := `{"response":"did-work","result":{"denied_actions":[{"display_name":"RunCommand"},{"display_name":"WriteFile"}]}}`
	scriptContent := "#!/bin/sh\necho \x27" + fakeJSON + "\x27\n"
	if err := os.WriteFile(script, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	stdout, _, exit, err := r.Run(worker.Job{Text: "job"})
	if err != nil || exit != 0 {
		t.Fatalf("run %v exit %d", err, exit)
	}
	if !strings.Contains(stdout, "did-work") {
		t.Fatalf("expected response in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "RunCommand") || !strings.Contains(stdout, "WriteFile") {
		t.Fatalf("expected denied actions in stdout, got %q", stdout)
	}
}

func TestAgyRunnerErrorAndStderrInReport(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	scriptContent := "#!/bin/sh\necho \x27error details on stderr\x27 >&2\necho \x27{\"status\":\"ERROR\",\"error\":\"agent failure message\"}\x27\n"
	if err := os.WriteFile(script, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	stdout, _, _, err := r.Run(worker.Job{Text: "job"})
	if err != nil {
		t.Fatalf("unexpected run err: %v", err)
	}
	if !strings.Contains(stdout, "agent failure message") {
		t.Fatalf("expected JSON error in report, got %q", stdout)
	}
	if !strings.Contains(stdout, "error details on stderr") {
		t.Fatalf("expected stderr in report, got %q", stdout)
	}
	if !strings.Contains(stdout, "status: ERROR") {
		t.Fatalf("expected status: ERROR in report, got %q", stdout)
	}
}
