package relay

import (
	"errors"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/dispatch"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestRelay_OnAgreed(t *testing.T) {
	fake := &worker.Fake{}
	r := &Relay{
		Worker: fake,
	}

	p := plan.Block{
		Worker: "JULES",
	}

	err := r.OnAgreed(p, "AGY", worker.Capabilities{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "NACK") {
		t.Errorf("expected error to contain NACK, got %q", err.Error())
	}
	if fake.Starts != 0 {
		t.Errorf("expected 0 Starts, got %d", fake.Starts)
	}
}

func TestRelay_OnRun_Asked(t *testing.T) {
	council := &state.Council{}
	fake := &worker.Fake{}
	r := &Relay{
		Council: council,
		Worker:  fake,
	}

	run := worker.Run{Outcome: worker.OutcomeAsked}
	act := r.OnRun(run)

	if act != dispatch.ActionAskSahil {
		t.Errorf("expected ActionAskSahil, got %v", act)
	}
	if council.Phase != state.PhaseWorkerAsked {
		t.Errorf("expected PhaseWorkerAsked, got %v", council.Phase)
	}
}

func TestRelay_OnRun_Failed(t *testing.T) {
	for _, outcome := range []worker.Outcome{worker.OutcomeFailed, worker.OutcomeTimedOut} {
		council := &state.Council{}
		fake := &worker.Fake{}
		r := &Relay{
			Council: council,
			Worker:  fake,
		}

		run := worker.Run{Outcome: outcome}
		act := r.OnRun(run)

		if act != dispatch.ActionFix {
			t.Errorf("%v: expected ActionFix, got %v", outcome, act)
		}
		if council.FixRounds != 0 {
			t.Errorf("%v: expected 0 FixRounds, got %d", outcome, council.FixRounds)
		}
		if council.Phase != state.PhaseFixing {
			t.Errorf("%v: expected PhaseFixing, got %v", outcome, council.Phase)
		}
	}
}

func TestRelay_OnRun_Cancelled(t *testing.T) {
	council := &state.Council{
		Handle: state.Handle{PID: 123},
	}
	fake := &worker.Fake{}
	r := &Relay{
		Council: council,
		Worker:  fake,
	}

	run := worker.Run{Outcome: worker.OutcomeCancelled}
	act := r.OnRun(run)

	if act != dispatch.ActionCleanup {
		t.Errorf("expected ActionCleanup, got %v", act)
	}
	if fake.Cancels != 1 {
		t.Errorf("expected 1 Cancel, got %d", fake.Cancels)
	}
	if !council.Handle.Zero() {
		t.Errorf("expected zero Handle, got %+v", council.Handle)
	}
}

func TestRelay_OnRun_Done(t *testing.T) {
	council := &state.Council{}
	fake := &worker.Fake{}
	r := &Relay{
		Council: council,
		Worker:  fake,
	}

	run := worker.Run{Outcome: worker.OutcomeDone}
	act := r.OnRun(run)

	if act != dispatch.ActionCollectEvidence {
		t.Errorf("expected ActionCollectEvidence, got %v", act)
	}
	if fake.Cancels != 0 {
		t.Errorf("expected 0 Cancels, got %d", fake.Cancels)
	}
}

func TestRelay_OnSIGHUP(t *testing.T) {
	council := &state.Council{
		Handle: state.Handle{PID: 123},
	}
	expectedErr := errors.New("cancel failed")
	fake := &worker.Fake{
		CancelErr: expectedErr,
	}
	r := &Relay{
		Council: council,
		Worker:  fake,
	}

	err := r.OnSIGHUP()
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if fake.Cancels != 1 {
		t.Errorf("expected 1 Cancel, got %d", fake.Cancels)
	}
	if !council.Handle.Zero() {
		t.Errorf("expected zero Handle, got %+v", council.Handle)
	}
}
