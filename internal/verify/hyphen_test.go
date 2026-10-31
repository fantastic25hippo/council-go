package verify

import (
	"os"
	"strings"
	"testing"
)

func TestRun_HyphenCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "council-verify-hyphen")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// A command starting with a hyphen (like -s or -e) should be treated
	// as a command to execute, not parsed as an option to bash.
	output, exit, err := Run(tempDir, "-s", 60)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	// exit should be 127 (bash: command not found) rather than bash syntax/option error
	if exit != 127 {
		t.Errorf("Run() exit = %d, want 127; output: %s", exit, output)
	}
	if !strings.Contains(output, "command not found") {
		t.Errorf("Run() output = %q, want containing 'command not found'", output)
	}
}
