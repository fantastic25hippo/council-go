1.  **Create `internal/state/run.go`**:
    *   Define `Handle` struct (`PID int`, `PGID int`, `SessionID string`).
    *   Implement `func (h Handle) Zero() bool`.
    *   Define `Council` struct (`Phase Phase`, `WorkerID string`, `Handle Handle`, `FixRounds int`, `LastOutcome worker.Outcome`).
    *   Implement `func (c *Council) SetRunning(workerID string, h Handle)`.
    *   Implement `func (c *Council) SetAsked()`.
    *   Implement `func (c *Council) SetFixing()`.
    *   Implement `func (c *Council) SetOutcome(o worker.Outcome)`.
    *   Implement `func (c *Council) ClearRun()`.
2.  **Create `internal/state/run_test.go`**:
    *   Write tests for `Handle{}.Zero()`, `Handle{PID:1}.Zero()`, `Handle{SessionID:"s"}.Zero()`.
    *   Write tests for `SetRunning`.
    *   Write tests for `SetAsked`.
    *   Write tests for `SetFixing` (twice -> FixRounds 2 and Phase FIXING).
    *   Write tests for `ClearRun`.
    *   Write tests for `SetOutcome`.
3.  **Run Tests**:
    *   Execute `go test ./internal/state/...` to ensure all tests pass.
4.  **Complete pre commit steps**:
    *   Ensure proper testing, verification, review, and reflection are done.
5.  **Submit Changes**:
    *   Create a PR against `master`.
