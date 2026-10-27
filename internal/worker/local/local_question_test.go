package local

import (
	"errors"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestWorker_QuestionWinsOverError(t *testing.T) {
	runner := &mockRunner{
		stdout: "some error logs\nCOUNCIL-QUESTION: which db?",
		exit:   1,
		err:    errors.New("command failed with exit 1"),
	}
	w := New(runner)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeAsked {
		t.Fatalf("Expected OutcomeAsked, got %v", run.Outcome)
	}
	if run.Question != "which db?" {
		t.Fatalf("Expected question 'which db?', got %q", run.Question)
	}
}

func TestWorker_NoQuestionWithError(t *testing.T) {
	runner := &mockRunner{
		stdout: "some error logs without question",
		exit:   1,
		err:    errors.New("command failed with exit 1"),
	}
	w := New(runner)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeFailed {
		t.Fatalf("Expected OutcomeFailed, got %v", run.Outcome)
	}
	if run.Question != "" {
		t.Fatalf("Expected empty question, got %q", run.Question)
	}
}
