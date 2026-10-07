package dispatch

import (
	"github.com/fantastic25hippo/council-go/internal/worker"
)

type Action int

const (
	ActionUnspecified Action = iota
	ActionAskSahil
	ActionFix
	ActionCleanup
	ActionCollectEvidence
	ActionHardStop
)

func (a Action) String() string {
	switch a {
	case ActionUnspecified:
		return "UNSPECIFIED"
	case ActionAskSahil:
		return "ASK_SAHIL"
	case ActionFix:
		return "FIX"
	case ActionCleanup:
		return "CLEANUP"
	case ActionCollectEvidence:
		return "COLLECT_EVIDENCE"
	case ActionHardStop:
		return "HARD_STOP"
	default:
		return "HARD_STOP"
	}
}

// Handle chooses the next relay action from run.Outcome only.
func Handle(run worker.Run) Action {
	switch run.Outcome {
	case worker.OutcomeAsked:
		return ActionAskSahil
	case worker.OutcomeFailed:
		return ActionFix
	case worker.OutcomeTimedOut:
		return ActionFix
	case worker.OutcomeCancelled:
		return ActionCleanup
	case worker.OutcomeDone:
		return ActionCollectEvidence
	default:
		return ActionHardStop
	}
}
