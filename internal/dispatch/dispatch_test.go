package dispatch

import (
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name     string
		run      worker.Run
		expected Action
	}{
		{
			name: "OutcomeAsked",
			run: worker.Run{
				Outcome: worker.OutcomeAsked,
			},
			expected: ActionAskSahil,
		},
		{
			name: "OutcomeFailed",
			run: worker.Run{
				Outcome: worker.OutcomeFailed,
			},
			expected: ActionFix,
		},
		{
			name: "OutcomeTimedOut",
			run: worker.Run{
				Outcome: worker.OutcomeTimedOut,
			},
			expected: ActionFix,
		},
		{
			name: "OutcomeCancelled",
			run: worker.Run{
				Outcome: worker.OutcomeCancelled,
			},
			expected: ActionCleanup,
		},
		{
			name: "OutcomeDone",
			run: worker.Run{
				Outcome: worker.OutcomeDone,
			},
			expected: ActionCollectEvidence,
		},
		{
			name: "OutcomeUnspecified",
			run: worker.Run{
				Outcome: worker.OutcomeUnspecified,
			},
			expected: ActionHardStop,
		},
		{
			name: "OutcomeDone with Report containing COUNCIL-QUESTION",
			run: worker.Run{
				Outcome: worker.OutcomeDone,
				Report:  "COUNCIL-QUESTION: wait?",
			},
			expected: ActionCollectEvidence,
		},
		{
			name: "OutcomeAsked with empty Question",
			run: worker.Run{
				Outcome:  worker.OutcomeAsked,
				Question: "",
			},
			expected: ActionAskSahil,
		},
		{
			name: "unknown Outcome",
			run: worker.Run{
				Outcome: worker.Outcome(99),
			},
			expected: ActionHardStop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Handle(tt.run)
			if actual != tt.expected {
				t.Errorf("Handle() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestActionString(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionUnspecified, "UNSPECIFIED"},
		{ActionAskSahil, "ASK_SAHIL"},
		{ActionFix, "FIX"},
		{ActionCleanup, "CLEANUP"},
		{ActionCollectEvidence, "COLLECT_EVIDENCE"},
		{ActionHardStop, "HARD_STOP"},
		{Action(99), "HARD_STOP"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			actual := tt.action.String()
			if actual != tt.expected {
				t.Errorf("Action.String() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}
