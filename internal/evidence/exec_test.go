package evidence

import (
	"strings"
	"testing"
)

func TestExecGH_ExitError(t *testing.T) {
	runner := DefaultGHRunner()
	stdout, _, exit, err := runner.Run([]string{"sh", "-c", "echo hello; exit 8"})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if exit != 8 {
		t.Errorf("expected exit 8, got %d", exit)
	}
	if !strings.Contains(stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got %q", stdout)
	}
}

func TestExecGH_CommandNotFound(t *testing.T) {
	runner := DefaultGHRunner()
	_, _, exit, err := runner.Run([]string{"nonexistent-command-xyz-12345"})
	if err == nil {
		t.Fatalf("expected non-nil error for nonexistent command")
	}
	if exit == 0 {
		t.Errorf("expected non-zero exit code when command fails to start, got %d", exit)
	}
}
