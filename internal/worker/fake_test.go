package worker

import "testing"

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
