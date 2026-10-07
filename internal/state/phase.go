package state

type Phase int

const (
	PhaseUnspecified Phase = iota
	PhaseWaitHI
	PhaseWaitHello
	PhaseDiscuss
	PhaseWaitAgreed
	PhaseWorkerRunning
	PhaseWorkerAsked
	PhaseFixing
	PhaseVerify
	PhaseReview
	PhaseAdjourn
)

func (p Phase) String() string {
	switch p {
	case PhaseWaitHI:
		return "WAIT_HI"
	case PhaseWaitHello:
		return "WAIT_HELLO"
	case PhaseDiscuss:
		return "DISCUSS"
	case PhaseWaitAgreed:
		return "WAIT_AGREED"
	case PhaseWorkerRunning:
		return "WORKER_RUNNING"
	case PhaseWorkerAsked:
		return "WORKER_ASKED"
	case PhaseFixing:
		return "FIXING"
	case PhaseVerify:
		return "VERIFY"
	case PhaseReview:
		return "REVIEW"
	case PhaseAdjourn:
		return "ADJOURN"
	case PhaseUnspecified:
		fallthrough
	default:
		return "UNSPECIFIED"
	}
}
