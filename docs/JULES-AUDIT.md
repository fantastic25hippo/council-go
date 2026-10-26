# Jules Audit Report

**Timeout (jules_timeout_secs)**
- The default worker timeout is 1800 seconds (30 minutes). See `internal/worker/jules/live.go:32`.
- It enforces this timeout by calculating a deadline `nowFn().Add(l.Timeout)` in `internal/worker/jules/live_poll.go:34`.
- The `http.Client` itself uses a default timeout of 180 seconds. See `internal/worker/jules/http.go:23`.

**What happens on timeout or on a too-big plan**
- If the poll loop hits the deadline, it returns `ErrPollTimeout` (`internal/worker/jules/live_poll.go:123`).
- The worker `Start()` catches `ErrPollTimeout` and translates it to a `worker.Run` with `worker.OutcomeTimedOut` (`internal/worker/jules/jules.go:44`).
- It does not process fixes in rounds for an oversized plan because the Jules worker currently blocks in `poll()` and waits for completion of the entire job on the Jules side.
- Plans cannot be sliced automatically by the Go worker code because the worker relies on `COMPLETED`, `FAILED`, `AWAITING_USER_FEEDBACK`, `PAUSED`, or `CANCELLED` states from the API to return (`internal/worker/jules/live_poll.go:102`).

**Single points of failure**
- A large plan can silently run out of time on the server-side, and the local worker will simply fail with a timeout `ErrPollTimeout` at 30 minutes without returning partial results.
- If `COMPLETED` is reached but the PR was not created during the session, it relies on a fallback PR logic (`l.fallbackPR()`) which could fail silently or return `nil` if PR fails to be pushed (`internal/worker/jules/live_poll.go:104`).
