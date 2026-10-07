package worker

import (
	"errors"
	"testing"
)

var _ Worker = (*Fake)(nil)

func TestFakeStartAndContinueRecordCalls(t *testing.T) {
	f := &Fake{Next: Run{Outcome: OutcomeDone, Report: "ok"}}
	job := Job{Text: "task"}
	got := f.Start(job)
	if got.Outcome != OutcomeDone || f.Starts != 1 {
		t.Fatalf("Start: outcome=%s starts=%d", got.Outcome, f.Starts)
	}
	f.Next = Run{Outcome: OutcomeAsked, Question: "which?"}
	got = f.Continue(job, "use local")
	if got.Outcome != OutcomeAsked || f.Continues != 1 || f.LastAnswer != "use local" {
		t.Fatalf("Continue: %+v continues=%d answer=%q", got, f.Continues, f.LastAnswer)
	}
}

func TestFakeCancel(t *testing.T) {
	errExpected := errors.New("cancel error")
	f := &Fake{CancelErr: errExpected}
	err := f.Cancel()
	if err != errExpected {
		t.Fatalf("Cancel: expected err=%v, got=%v", errExpected, err)
	}
	if f.Cancels != 1 {
		t.Fatalf("Cancel: expected cancels=1, got=%d", f.Cancels)
	}
}
