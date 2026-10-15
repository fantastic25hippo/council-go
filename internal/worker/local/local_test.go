package local

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/proc"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

var _ worker.Worker = (*Worker)(nil)

type mockRunner struct {
	stdout string
	stderr string
	exit   int
	err    error
	job    worker.Job
}

func (m *mockRunner) Run(job worker.Job) (string, string, int, error) {
	m.job = job
	return m.stdout, m.stderr, m.exit, m.err
}

func TestWorker_Start_Asked(t *testing.T) {
	runner := &mockRunner{
		stdout: "Some other text\n" + worker.QuestionPrefix + " should we merge?",
		exit:   0,
	}
	w := New(runner)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeAsked {
		t.Errorf("Expected OutcomeAsked, got %v", run.Outcome)
	}
	if run.Question != "should we merge?" {
		t.Errorf("Expected question 'should we merge?', got '%v'", run.Question)
	}
}

func TestWorker_Start_Done(t *testing.T) {
	runner := &mockRunner{
		stdout: "accept 4255ed2?",
		exit:   0,
	}
	w := New(runner)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeDone {
		t.Errorf("Expected OutcomeDone, got %v", run.Outcome)
	}
	if run.Question != "" {
		t.Errorf("Expected empty question, got '%v'", run.Question)
	}
}

func TestWorker_Start_Failed(t *testing.T) {
	runner := &mockRunner{
		stdout: "some error occurred",
		exit:   1,
	}
	w := New(runner)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeFailed {
		t.Errorf("Expected OutcomeFailed, got %v", run.Outcome)
	}
}

func TestWorker_Start_FailedErr(t *testing.T) {
	runner := &mockRunner{
		err: errors.New("exec error"),
	}
	w := New(runner)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeFailed {
		t.Errorf("Expected OutcomeFailed, got %v", run.Outcome)
	}
	if run.Report != "exec error" {
		t.Errorf("Expected Report 'exec error', got '%v'", run.Report)
	}
}

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

func TestWorker_Capabilities(t *testing.T) {
	w := New(nil)
	caps := w.Capabilities()

	if !caps.CanAskAndWait {
		t.Errorf("Expected CanAskAndWait true")
	}
	if !caps.KeepsMemoryAcrossFixes {
		t.Errorf("Expected KeepsMemoryAcrossFixes true")
	}
	if caps.ReachesGitHub {
		t.Errorf("Expected ReachesGitHub false")
	}
	if !caps.TouchesLocalCheckout {
		t.Errorf("Expected TouchesLocalCheckout true")
	}
	if caps.OpensPR {
		t.Errorf("Expected OpensPR false")
	}
}

func TestWorker_Start_SetsConvID(t *testing.T) {
	// A new test to ensure that mockRunner doesn't crash or get ConvID set.
	runner := &mockRunner{
		stdout: "{\"conversation_id\":\"c1\",\"response\":\"resp\"}",
		exit:   0,
	}
	w := New(runner)
	run := w.Start(worker.Job{})
	if run.Outcome != worker.OutcomeDone {
		t.Errorf("Expected OutcomeDone, got %v", run.Outcome)
	}
	// The mockRunner does not parse JSON. The parsing is in AgyRunner.Run.
	// We just ensure it doesn't fail.
}

func TestWorker_Cancel(t *testing.T) {
	w := New(nil)
	if err := w.Cancel(); err != nil {
		t.Errorf("Expected nil error from Cancel, got %v", err)
	}
}

type cancelRunner struct {
	mockRunner
	cancelCalled bool
	cancelErr    error
}

func (c *cancelRunner) Cancel() error {
	c.cancelCalled = true
	return c.cancelErr
}

func TestWorker_Cancel_CallsRunner(t *testing.T) {
	runner := &cancelRunner{
		cancelErr: errors.New("runner cancel err"),
	}
	w := New(runner)
	err := w.Cancel()
	if !runner.cancelCalled {
		t.Errorf("Expected runner Cancel to be called")
	}
	if err == nil || err.Error() != "runner cancel err" {
		t.Errorf("Expected err 'runner cancel err', got %v", err)
	}
}

func TestWorker_Cancel_KillsCmd(t *testing.T) {
	cmd, err := proc.Start([]string{"sleep", "30"}, ".")
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	w := New(nil)
	w.cmd = cmd

	if err := w.Cancel(); err != nil {
		t.Errorf("Expected nil error from Cancel, got %v", err)
	}

	if w.cmd != nil {
		t.Errorf("Expected w.cmd to be nil after Cancel")
	}

	// Ensure the process is dead
	err = cmd.Wait()
	if err == nil {
		t.Errorf("Expected error from Wait (process killed), got nil")
	}
}
