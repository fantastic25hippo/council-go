package proc_test

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/proc"
)

func TestStartEmptyArgv(t *testing.T) {
	_, err := proc.Start([]string{}, "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "argv") {
		t.Fatalf("expected error to contain 'argv', got: %v", err)
	}
}

func TestKillNilCmd(t *testing.T) {
	err := proc.Kill(nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "process") {
		t.Fatalf("expected error to contain 'process', got: %v", err)
	}
}

func TestKillNilProcess(t *testing.T) {
	err := proc.Kill(&exec.Cmd{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "process") {
		t.Fatalf("expected error to contain 'process', got: %v", err)
	}
}

func TestStartAndKill(t *testing.T) {
	tempdir := t.TempDir()

	cmd, err := proc.Start([]string{"sleep", "30"}, tempdir)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = proc.Kill(cmd)
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("expected an error because process was killed")
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Wait blocked for more than 3 seconds")
	}
}

func TestProcessGroupKillAllowed(t *testing.T) {
    tempdir := t.TempDir()

    cmd, err := proc.Start([]string{"sleep", "30"}, tempdir)
    if err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    err = cmd.Process.Kill()
    if err != nil {
        t.Fatalf("cmd.Process.Kill failed: %v", err)
    }

    _ = cmd.Wait()
}
