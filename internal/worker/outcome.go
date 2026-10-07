package worker

// Outcome is the only control-flow result of a worker run.
type Outcome int

const (
	OutcomeUnspecified Outcome = iota
	OutcomeDone
	OutcomeAsked
	OutcomeFailed
	OutcomeTimedOut
	OutcomeCancelled
)

func (o Outcome) String() string {
	switch o {
	case OutcomeDone:
		return "DONE"
	case OutcomeAsked:
		return "ASKED"
	case OutcomeFailed:
		return "FAILED"
	case OutcomeTimedOut:
		return "TIMED_OUT"
	case OutcomeCancelled:
		return "CANCELLED"
	default:
		return "UNSPECIFIED"
	}
}
