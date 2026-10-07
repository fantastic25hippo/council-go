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

func agreedPlan() string {
	return `WORKER: local
OWNED_PATHS:
- internal/foo
CHANGE: add a thing
VERIFY: echo hi
AGREED`
}

func TestRunQuitAtHi(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok"}
	b := &member.Fake{Name: "gemini", Shown: "Gemini"}
	fakeW := &worker.Fake{Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true}}
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:       cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:         a,
		B:         b,
		IO:        &term.IO{In: strings.NewReader("quit\n"), Out: &bytes.Buffer{}},
		Council:   council,
		Relay:     &relay.Relay{Council: council, Worker: fakeW},
		WorkerFor: func(string) worker.Worker { return fakeW },
		Cwd:       t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if council.Phase != state.PhaseAdjourn {
		t.Fatalf("phase %s", council.Phase)
	}
	if a.Closed != 1 || b.Closed != 1 {
		t.Fatalf("close a=%d b=%d", a.Closed, b.Closed)
	}
	if fakeW.Starts != 0 {
		t.Fatalf("worker started")
	}
}

func TestRunAgreeWorkPass(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS\n- looks good",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: PASS\n- ok",
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
		IO:      &term.IO{In: strings.NewReader("hi\nfix the foo\n\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: fakeW},
		WorkerFor: func(id string) worker.Worker {
			if id != "local" {
				t.Errorf("worker id %s", id)
			}
			return fakeW
		},
		Cwd: t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if council.Phase != state.PhaseAdjourn {
		t.Fatalf("phase %s out=%s", council.Phase, out.String())
	}
	if fakeW.Starts != 1 {
		t.Fatalf("starts %d", fakeW.Starts)
	}
	if a.Started != 1 || b.Started != 1 {
		t.Fatalf("started a=%d b=%d", a.Started, b.Started)
	}
	if !strings.Contains(out.String(), "Dual AGREED") {
		t.Fatalf("output %s", out.String())
	}
	if !strings.Contains(out.String(), "Both PASS") {
		t.Fatalf("no pass %s", out.String())
	}
}
