package state

import (
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestHandleZero(t *testing.T) {
	if !(Handle{}).Zero() {
		t.Error("expected Handle{}.Zero() to be true")
	}
	if (Handle{PID: 1}).Zero() {
		t.Error("expected Handle{PID: 1}.Zero() to be false")
	}
	if (Handle{SessionID: "s"}).Zero() {
		t.Error(`expected Handle{SessionID: "s"}.Zero() to be false`)
	}
}

func TestSetRunning(t *testing.T) {
	c := &Council{}
	c.SetRunning("local", Handle{PID: 9, PGID: 9})
	if c.Phase != PhaseWorkerRunning {
		t.Errorf("expected Phase to be %v, got %v", PhaseWorkerRunning, c.Phase)
	}
	if c.WorkerID != "local" {
		t.Errorf("expected WorkerID to be local, got %v", c.WorkerID)
	}
	if c.Handle.PID != 9 || c.Handle.PGID != 9 {
		t.Errorf("expected Handle to be {PID: 9, PGID: 9}, got %v", c.Handle)
	}
}

func TestSetAsked(t *testing.T) {
	c := &Council{
		Handle: Handle{PID: 1},
	}
	c.SetAsked()
	if c.Phase != PhaseWorkerAsked {
		t.Errorf("expected Phase to be %v, got %v", PhaseWorkerAsked, c.Phase)
	}
	if c.Handle.PID != 1 {
		t.Errorf("expected Handle to be preserved, got %v", c.Handle)
	}
}

func TestSetFixing(t *testing.T) {
	c := &Council{}
	c.SetFixing()
	c.SetFixing()
	if c.FixRounds != 2 {
		t.Errorf("expected FixRounds to be 2, got %v", c.FixRounds)
	}
	if c.Phase != PhaseFixing {
		t.Errorf("expected Phase to be %v, got %v", PhaseFixing, c.Phase)
	}
}

func TestClearRun(t *testing.T) {
	c := &Council{
		Phase:  PhaseWorkerRunning,
		Handle: Handle{PID: 1},
	}
	c.ClearRun()
	if !c.Handle.Zero() {
		t.Errorf("expected Handle to be zero, got %v", c.Handle)
	}
	if c.Phase != PhaseWorkerRunning {
		t.Errorf("expected Phase to remain unchanged, got %v", c.Phase)
	}
}

func TestSetOutcome(t *testing.T) {
	c := &Council{}
	c.SetOutcome(worker.OutcomeFailed)
	if c.LastOutcome != worker.OutcomeFailed {
		t.Errorf("expected LastOutcome to be %v, got %v", worker.OutcomeFailed, c.LastOutcome)
	}
}
