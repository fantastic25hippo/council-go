package meeting_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/meeting"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func splitSession(t *testing.T, a, b *member.Fake, w *worker.Fake, in string) (*meeting.Session, string) {
	t.Helper()
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader(in), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: w},
		WorkerFor: func(string) worker.Worker {
			return w
		},
		Cwd: t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s, out.String()
}

func sentSplit(m *member.Fake) bool {
	for _, p := range m.Sent {
		if strings.Contains(p, "You may keep or change your verdict") {
			return true
		}
	}
	return false
}

func TestRunSplitThenAgree(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS\n- looks good",
		"VERDICT: PASS\n- ok after seeing theirs",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL\n- missing tests",
		"VERDICT: PASS\n- ok now",
	}}
	w := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
	}
	s, out := splitSession(t, a, b, w, "hi\nfix the foo\n\nquit\n")
	if s.Council.Phase != state.PhaseAdjourn {
		t.Fatalf("phase %s out=%s", s.Council.Phase, out)
	}
	if w.Starts != 1 {
		t.Fatalf("starts %d", w.Starts)
	}
	if w.Continues != 0 {
		t.Fatalf("continues %d", w.Continues)
	}
	if !strings.Contains(out, "Both PASS") {
		t.Fatalf("no pass %s", out)
	}
	if !sentSplit(a) || !sentSplit(b) {
		t.Fatalf("expected SplitReview")
	}
}

func TestRunSplitAskSahilQuit(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS\n- looks good",
		"VERDICT: PASS\n- still pass",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL\n- missing tests",
		"VERDICT: FAIL\n- still fail",
	}}
	w := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
	}
	s, out := splitSession(t, a, b, w, "hi\nfix the foo\n\nquit\n")
	if s.Council.Phase != state.PhaseAdjourn {
		t.Fatalf("phase %s out=%s", s.Council.Phase, out)
	}
	if w.Starts != 1 {
		t.Fatalf("starts %d", w.Starts)
	}
	if w.Continues != 0 {
		t.Fatalf("continues %d want 0", w.Continues)
	}
	if !strings.Contains(out, "Split verdicts") {
		t.Fatalf("no split prompt %s", out)
	}
}

func TestRunBothFailSkipsSplitReview(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: FAIL\n- broken",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL\n- broken",
	}}
	w := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
	}
	w.HandleContinue = func(worker.Job, string) worker.Run {
		return worker.Run{Outcome: worker.OutcomeUnspecified, Report: "stop"}
	}
	w.Next = worker.Run{Outcome: worker.OutcomeDone, Report: "did it"}
	_, out := splitSession(t, a, b, w, "hi\nfix the foo\n\nquit\n")
	if w.Starts != 1 {
		t.Fatalf("starts %d", w.Starts)
	}
	if w.Continues != 1 {
		t.Fatalf("continues %d", w.Continues)
	}
	if sentSplit(a) || sentSplit(b) {
		t.Fatalf("SplitReview on dual FAIL out=%s", out)
	}
}
