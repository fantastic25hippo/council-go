package meeting_test

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestRunSplitAskSahilPass(t *testing.T) {
	for _, cmd := range []string{"pass", "PASS"} {
		t.Run(cmd, func(t *testing.T) {
			a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
				"Hello, I am Grok.",
				agreedPlan(),
				"VERDICT: PASS\n- looks good",
				"VERDICT: PASS\n- still pass",
				"Sahil accepted.\nSAHIL_MEANS: PASS",
			}}
			b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
				"Hello, I am Gemini.",
				"local is fine",
				"LAUNCH: local\nAGREED",
				"VERDICT: FAIL\n- missing tests",
				"VERDICT: FAIL\n- still fail",
				"Sahil approved.\nSAHIL_MEANS: PASS",
			}}
			w := &worker.Fake{
				Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
				Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
			}
			s, out := splitSession(t, a, b, w, "hi\nfix the foo\n\n"+cmd+"\nquit\n")
			if s.Council.Phase != state.PhaseAdjourn {
				t.Fatalf("phase %s out=%s", s.Council.Phase, out)
			}
			if w.Starts != 1 {
				t.Fatalf("starts %d want 1", w.Starts)
			}
			if w.Continues != 0 {
				t.Fatalf("continues %d want 0", w.Continues)
			}
			if !strings.Contains(out, "Both PASS") {
				t.Fatalf("no Both PASS in out: %s", out)
			}
		})
	}
}

func TestRunSplitAskSahilFail(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS\n- looks good",
		"VERDICT: PASS\n- still pass",
		"Sahil wants to fail it.\nSAHIL_MEANS: FAIL",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL\n- missing tests",
		"VERDICT: FAIL\n- still fail",
		"Sahil rejected.\nSAHIL_MEANS: FAIL",
	}}
	w := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
	}
	w.HandleContinue = func(worker.Job, string) worker.Run {
		return worker.Run{Outcome: worker.OutcomeUnspecified, Report: "stop"}
	}
	s, _ := splitSession(t, a, b, w, "hi\nfix the foo\n\nfail\nquit\n")
	if s.Council.Phase != state.PhaseAdjourn {
		t.Fatalf("phase %s", s.Council.Phase)
	}
	if w.Starts != 1 {
		t.Fatalf("starts %d want 1", w.Starts)
	}
	if w.Continues != 1 {
		t.Fatalf("continues %d want 1", w.Continues)
	}
	if !strings.Contains(w.LastAnswer, "still pass") || !strings.Contains(w.LastAnswer, "still fail") {
		t.Fatalf("expected latest verdicts in fix text, got: %s", w.LastAnswer)
	}
}
