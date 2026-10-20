package state

import (
	"sync"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type Handle struct {
	PID       int
	PGID      int
	SessionID string
}

func (h Handle) Zero() bool {
	return h.PID == 0 && h.PGID == 0 && h.SessionID == ""
}

type Council struct {
	mu          sync.RWMutex
	Phase       Phase
	WorkerID    string
	Handle      Handle
	FixRounds   int
	Asks        int
	LastOutcome worker.Outcome
	LastRun     worker.Run
}

func (c *Council) Snapshot() Council {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Council{
		Phase:       c.Phase,
		WorkerID:    c.WorkerID,
		Handle:      c.Handle,
		FixRounds:   c.FixRounds,
		Asks:        c.Asks,
		LastOutcome: c.LastOutcome,
		LastRun:     c.LastRun,
	}
}

func (c *Council) SetRunning(workerID string, h Handle) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Phase = PhaseWorkerRunning
	c.WorkerID = workerID
	c.Handle = h
}

func (c *Council) SetAsked() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Phase = PhaseWorkerAsked
}

func (c *Council) SetFixing() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FixRounds++
	c.Phase = PhaseFixing
}

func (c *Council) SetOutcome(o worker.Outcome) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastOutcome = o
}

func (c *Council) SetRun(r worker.Run) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastRun = r
	c.LastOutcome = r.Outcome
}

func (c *Council) ClearRun() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Handle = Handle{}
}

func (c *Council) GetPhase() Phase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Phase
}

func (c *Council) GetLastRun() worker.Run {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastRun
}
