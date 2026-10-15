package prompt

import "strings"

import "testing"

import (
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestSystemAndGreet(t *testing.T) {
	s := System("Grok", "Gemini", 3, 3, []string{"local", "jules"})
	if !strings.Contains(s, "Grok") || !strings.Contains(s, "AGREED") || strings.Contains(s, "python3") {
		t.Fatalf("system %q", s)
	}
	g := Greet("Grok", "Gemini")
	if !strings.Contains(g, "Grok") || strings.Contains(strings.ToLower(g), "propose") && strings.Contains(g, "Do not start proposing") {
		// must contain do not start proposing
	}
	if !strings.Contains(g, "Do not start proposing") {
		t.Fatalf("greet %q", g)
	}
}

func TestSplitReviewNamesOther(t *testing.T) {
	s := SplitReview("Gemini", "VERDICT: FAIL\n- tests")
	if !strings.Contains(s, "Gemini's verdict:") || !strings.Contains(s, "VERDICT: FAIL") {
		t.Fatalf("%q", s)
	}
}

func TestReviewAskHasVerdictInstruction(t *testing.T) {
	s := ReviewAsk("d", "out", 0, "claim")
	if !strings.Contains(s, "VERDICT: PASS") || strings.Contains(s, "python3") {
		t.Fatalf("%q", s)
	}
}

func TestProposalHasResolvedWorker(t *testing.T) {
	s := Proposal("my task", []string{"local", "jules"}, "local", worker.Capabilities{CanAskAndWait: true})
	if !strings.Contains(s, "This task will be run by local") {
		t.Fatalf("expected resolved worker string in %q", s)
	}
	if !strings.Contains(s, "CanAskAndWait: true") || !strings.Contains(s, "ReachesGitHub: false") {
		t.Fatalf("expected capabilities strings in %q", s)
	}
}

func TestNackPlanReason(t *testing.T) {
	s := NackPlanReason([]string{"local", "jules"}, "protected path overlap")
	if !strings.Contains(s, "protected path overlap") || !strings.Contains(s, "local, jules") || !strings.Contains(s, "Resend the full plan") {
		t.Fatalf("unexpected NackPlanReason: %q", s)
	}
}
