package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestWorker_Continue(t *testing.T) {
	runner := &mockRunner{
		stdout: "success",
		exit:   0,
	}
	w := New(runner)

	originalJob := worker.Job{
		Text: "some original text\n### VERIFICATION FAILURE\nfailed stuff",
	}

	w.Continue(originalJob, "my answer")

	if !strings.Contains(runner.job.Text, "my answer") {
		t.Errorf("Expected Text to contain answer, got '%v'", runner.job.Text)
	}
}

func TestWorker_Continue_WithConvID(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	content := "#!/bin/sh\neval echo \"\\$$#\"\n"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	runner := &AgyRunner{
		ConvID:  "c1",
		CLI:     script,
		Cwd:     dir,
		Timeout: 5 * time.Second,
	}
	w := New(runner)
	run := w.Continue(worker.Job{Text: "ORIGINAL JOB"}, "just the answer")
	if strings.TrimSpace(run.Report) != "just the answer" {
		t.Fatalf("expected report 'just the answer', got %q", run.Report)
	}
	if strings.Contains(run.Report, "ORIGINAL JOB") {
		t.Fatalf("report should not contain ORIGINAL JOB, got %q", run.Report)
	}
}
