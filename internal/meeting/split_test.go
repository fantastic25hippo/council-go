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

func TestRunSplitThenAgree(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS",
		"VERDICT: PASS",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL",
		"VERDICT: PASS",
	}}

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nfix the foo\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: fakeW},
		WorkerFor: func(id string) worker.Worker {
			return fakeW
		},
		Cwd: t.TempDir(),
	}

	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fakeW.Starts != 1 {
		t.Fatalf("expected 1 start, got %d", fakeW.Starts)
	}
	if fakeW.Continues != 0 {
		t.Fatalf("expected 0 continues, got %d", fakeW.Continues)
	}
	if !strings.Contains(out.String(), "Both PASS") {
		t.Fatalf("expected Both PASS, got %s", out.String())
	}
}

func TestRunSplitThenSahilQuit(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS",
		"VERDICT: PASS",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL",
		"VERDICT: FAIL",
	}}

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nfix the foo\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: fakeW},
		WorkerFor: func(id string) worker.Worker {
			return fakeW
		},
		Cwd: t.TempDir(),
	}

	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fakeW.Starts != 1 {
		t.Fatalf("expected 1 start, got %d", fakeW.Starts)
	}
	if fakeW.Continues != 0 {
		t.Fatalf("expected 0 continues, got %d", fakeW.Continues)
	}
	if council.Phase != state.PhaseAdjourn {
		t.Fatalf("expected phase adjourn, got %s", council.Phase)
	}
}

func TestRunBothFAILNoExtraRound(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: FAIL",
		"FIX", // FIX without extra round
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: FAIL",
		"FIX", // FIX without extra round
	}}

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
	}
	fakeW.HandleContinue = func(job worker.Job, answer string) worker.Run {
		return worker.Run{Outcome: worker.OutcomeUnspecified, Report: "done"} // mapped to HardStop
	}
	fakeW.Next = worker.Run{Outcome: worker.OutcomeDone, Report: "did it"} // DONE forces verdicts

	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nfix the foo\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: fakeW},
		WorkerFor: func(id string) worker.Worker {
			return fakeW
		},
		Cwd: t.TempDir(),
	}

	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fakeW.Starts != 1 {
		t.Fatalf("expected 1 start, got %d", fakeW.Starts)
	}
	if fakeW.Continues != 1 {
		t.Fatalf("expected 1 continue, got %d", fakeW.Continues)
	}
}
