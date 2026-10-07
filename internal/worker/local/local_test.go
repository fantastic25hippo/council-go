package local

import (
	"errors"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

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

	if !strings.Contains(runner.job.Text, "ANSWER:my answer") {
		t.Errorf("Expected Text to contain ANSWER, got '%v'", runner.job.Text)
	}
	if strings.Contains(runner.job.Text, "VERIFICATION FAILURE") {
		t.Errorf("Expected Text to not contain VERIFICATION FAILURE, got '%v'", runner.job.Text)
	}
}

func TestWorker_Capabilities(t *testing.T) {
	w := New(nil)
	caps := w.Capabilities()

	if !caps.CanAskAndWait {
		t.Errorf("Expected CanAskAndWait true")
	}
	if caps.KeepsMemoryAcrossFixes {
		t.Errorf("Expected KeepsMemoryAcrossFixes false")
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
