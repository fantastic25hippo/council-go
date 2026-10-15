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

func TestTaskDoneThenSecondTask(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		agreedPlan(),
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"local is fine",
		"LAUNCH: local\nAGREED",
	}}

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{
			TouchesLocalCheckout: true,
			CanAskAndWait:        true,
		},
		Next: worker.Run{Outcome: worker.OutcomeFailed, Report: "crash"},
	}

	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 0, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nfirst task\n\nsecond task\n\nquit\n"), Out: &out},
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

	if fakeW.Starts != 2 {
		t.Fatalf("expected 2 starts, got %d", fakeW.Starts)
	}
	if !strings.Contains(out.String(), "Fix round cap reached") {
		t.Fatalf("expected output to contain 'Fix round cap reached', got %s", out.String())
	}
	if council.Phase != state.PhaseAdjourn {
		t.Fatalf("expected phase %s, got %s", state.PhaseAdjourn, council.Phase)
	}
}

func TestNoWorkerThenSecondTask(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		agreedPlan(),
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"local is fine",
		"LAUNCH: local\nAGREED",
	}}

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{
			TouchesLocalCheckout: true,
			CanAskAndWait:        true,
		},
	}

	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nfirst task\n\nsecond task\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: fakeW},
		WorkerFor: func(id string) worker.Worker {
			return nil
		},
		Cwd: t.TempDir(),
	}

	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "no worker") {
		t.Fatalf("expected output to contain 'no worker', got %s", out.String())
	}
	if fakeW.Starts != 0 {
		t.Fatalf("expected 0 starts, got %d", fakeW.Starts)
	}
	if strings.Count(out.String(), "Council is ready") < 2 {
		t.Fatalf("expected 'Council is ready' at least twice, got %d", strings.Count(out.String(), "Council is ready"))
	}
	if council.Phase != state.PhaseAdjourn {
		t.Fatalf("expected phase %s, got %s", state.PhaseAdjourn, council.Phase)
	}
}
