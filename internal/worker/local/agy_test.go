package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestAgyRunnerRun(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho did-the-work\n"), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	stdout, _, exit, err := r.Run(worker.Job{Text: "job"})
	if err != nil || exit != 0 {
		t.Fatalf("run %v exit %d", err, exit)
	}
	if !strings.Contains(stdout, "did-the-work") {
		t.Fatalf("stdout %q", stdout)
	}
}

func TestAgyRunnerArgvPutsDashPLast(t *testing.T) {
	r := &AgyRunner{CLI: "agy", Model: "m", Cwd: t.TempDir()}
	argv := r.argv(worker.Job{Text: "hello", Paths: nil})
	if len(argv) < 2 || argv[len(argv)-2] != "-p" || argv[len(argv)-1] != "hello" {
		t.Fatalf("argv %v", argv)
	}

	// Should contain --output-format json before -p, and no --conversation
	foundJSON := false
	for i, a := range argv {
		if a == "--output-format" && i+1 < len(argv) && argv[i+1] == "json" {
			foundJSON = true
		}
		if a == "--conversation" {
			t.Fatalf("did not expect --conversation when ConvID is empty")
		}
	}
	if !foundJSON {
		t.Fatalf("argv missing --output-format json")
	}
}

func TestAgyRunnerArgvWithConvID(t *testing.T) {
	r := &AgyRunner{CLI: "agy", ConvID: "abc", Cwd: t.TempDir()}
	argv := r.argv(worker.Job{Text: "hello", Paths: nil})

	foundConv := false
	for i, a := range argv {
		if a == "--conversation" && i+1 < len(argv) && argv[i+1] == "abc" {
			foundConv = true
			if i > len(argv)-3 {
				t.Fatalf("--conversation must be before -p")
			}
		}
	}
	if !foundConv {
		t.Fatalf("argv missing --conversation abc")
	}
}

func TestAgyRunnerJSONParseTopLevel(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho '{\"conversation_id\":\"c1\",\"response\":\"did-the-work\"}'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	stdout, _, exit, err := r.Run(worker.Job{Text: "job"})
	if err != nil || exit != 0 {
		t.Fatalf("run %v exit %d", err, exit)
	}
	if stdout != "did-the-work" {
		t.Fatalf("expected stdout 'did-the-work', got %q", stdout)
	}
	if r.ConvID != "c1" {
		t.Fatalf("expected ConvID 'c1', got %q", r.ConvID)
	}
}

func TestAgyRunnerJSONParseNestedResult(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho '{\"result\":{\"conversation_id\":\"c2\",\"response\":\"ok\"}}'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 5 * time.Second}
	stdout, _, exit, err := r.Run(worker.Job{Text: "job"})
	if err != nil || exit != 0 {
		t.Fatalf("run %v exit %d", err, exit)
	}
	if stdout != "ok" {
		t.Fatalf("expected stdout 'ok', got %q", stdout)
	}
	if r.ConvID != "c2" {
		t.Fatalf("expected ConvID 'c2', got %q", r.ConvID)
	}
}

func TestAgyRunnerTimeout(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "sleep-agy")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\n"), 0755); err != nil {
		t.Fatal(err)
	}
	r := &AgyRunner{CLI: script, Cwd: dir, Timeout: 200 * time.Millisecond}
	_, _, _, err := r.Run(worker.Job{Text: "job"})
	if err != ErrTimeout {
		t.Fatalf("want ErrTimeout, got %v", err)
	}
}

func TestWorkerStartTimeoutOutcome(t *testing.T) {
	w := New(&timeoutRunner{})
	run := w.Start(worker.Job{})
	if run.Outcome != worker.OutcomeTimedOut {
		t.Fatalf("outcome %v", run.Outcome)
	}
}

type timeoutRunner struct{}

func (timeoutRunner) Run(worker.Job) (string, string, int, error) {
	return "", "", 1, ErrTimeout
}
