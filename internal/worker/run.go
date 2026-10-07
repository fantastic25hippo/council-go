package worker

// Job is the input to Start and Continue. The plan text is immutable
// after AGREED; adapters must not edit it.
type Job struct {
	Plan   string
	Text   string
	Paths  []string
	Verify string
	Repo   string
}

// Evidence is collected by the relay on DONE. Phase 1 leaves it empty.
type Evidence struct {
	Diff         string
	VerifyOutput string
	VerifyExit   int
}

// Raw is runtime-specific debug data. Never parse it to choose Outcome.
type Raw struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	SessionID string
	PRURL     string
}

// Run is the typed result of Start or Continue.
type Run struct {
	Outcome  Outcome
	Report   string
	Question string
	Evidence Evidence
	Raw      Raw
}
