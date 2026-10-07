package state_test

import (
	"os"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/state"
)

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase    state.Phase
		expected string
	}{
		{state.PhaseUnspecified, "UNSPECIFIED"},
		{state.PhaseWaitHI, "WAIT_HI"},
		{state.PhaseWaitHello, "WAIT_HELLO"},
		{state.PhaseDiscuss, "DISCUSS"},
		{state.PhaseWaitAgreed, "WAIT_AGREED"},
		{state.PhaseWorkerRunning, "WORKER_RUNNING"},
		{state.PhaseWorkerAsked, "WORKER_ASKED"},
		{state.PhaseFixing, "FIXING"},
		{state.PhaseVerify, "VERIFY"},
		{state.PhaseReview, "REVIEW"},
		{state.PhaseAdjourn, "ADJOURN"},
		{state.Phase(99), "UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.expected {
				t.Errorf("Phase.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNoWaitJules(t *testing.T) {
	content, err := os.ReadFile("phase.go")
	if err != nil {
		t.Fatalf("Failed to read phase.go: %v", err)
	}

	if strings.Contains(string(content), "WAIT_JULES") {
		t.Errorf("Source file contains forbidden substring 'WAIT_JULES'")
	}
}
