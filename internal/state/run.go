package state

import "github.com/fantastic25hippo/council-go/internal/worker"

type Handle struct {
	PID       int
	PGID      int
	SessionID string
}

func (h Handle) Zero() bool {
	return h.PID == 0 && h.PGID == 0 && h.SessionID == ""
}

type Council struct {
	Phase       Phase
	WorkerID    string
	Handle      Handle
	FixRounds   int
	LastOutcome worker.Outcome
	LastRun     worker.Run
}

func (c *Council) SetRunning(workerID string, h Handle) {
	c.Phase = PhaseWorkerRunning
	c.WorkerID = workerID
	c.Handle = h
}

func (c *Council) SetAsked() {
	c.Phase = PhaseWorkerAsked
}

func (c *Council) SetFixing() {
	c.FixRounds++
	c.Phase = PhaseFixing
}

func (c *Council) SetOutcome(o worker.Outcome) {
	c.LastOutcome = o
}

func (c *Council) SetRun(r worker.Run) {
	c.LastRun = r
	c.LastOutcome = r.Outcome
}

func (c *Council) ClearRun() {
	c.Handle = Handle{}
}
