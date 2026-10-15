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
