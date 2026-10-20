package local

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestWorkerCancelLongRunningCommand(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "sleep-cmd")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\n"), 0755); err != nil {
		t.Fatal(err)
	}

	runner := &AgyRunner{
		CLI:     script,
		Cwd:     dir,
		Timeout: 10 * time.Second,
	}
	w := New(runner)

	done := make(chan worker.Run, 1)
	go func() {
		done <- w.Start(worker.Job{Text: "run"})
	}()

	time.Sleep(50 * time.Millisecond)

	if err := w.Cancel(); err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}

	select {
	case run := <-done:
		if run.Outcome != worker.OutcomeCancelled {
			t.Fatalf("expected OutcomeCancelled, got %v (report: %q, exit: %d)", run.Outcome, run.Report, run.Raw.ExitCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker to cancel")
	}
}
