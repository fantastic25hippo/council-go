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

func planWithPaths(paths ...string) string {
	s := "WORKER: local\nOWNED_PATHS:\n"
	for _, p := range paths {
		s += "- " + p + "\n"
	}
	s += "CHANGE: add a thing\nVERIFY: echo hi\nAGREED"
	return s
}

func TestAgreeDualNackDifferentOwnedPathsNotAdopted(t *testing.T) {
	planA := planWithPaths("internal/foo")
	planB := planWithPaths("internal/bar")

	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		protectedPlan(),
		planA,
		"debate grok",
		"disputed grok",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"ok",
		"LAUNCH: local\nAGREED",
		planB,
		"debate gemini",
		"disputed gemini",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "done"},
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 1, FixRoundsMax: 3, Worker: "local"},
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

	if fakeW.Starts != 0 {
		t.Fatalf("expected worker NOT to start when plans have different OwnedPaths, got %d starts", fakeW.Starts)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "Members still failed to provide a valid plan.") {
		t.Fatalf("expected failure message in output, got: %s", outStr)
	}
}

func TestAgreeDualNackEqualOwnedPathsAdopted(t *testing.T) {
	planA := "WORKER: local\nOWNED_PATHS:\n- internal/foo\n- internal/bar\nCHANGE: add a thing\nVERIFY: echo hi\nAGREED"
	planB := "WORKER: local\nOWNED_PATHS:\n-  internal/bar \n- internal/foo\nCHANGE:   add a thing  \nVERIFY: echo hi\nAGREED"

	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		protectedPlan(),
		planA,
		"VERDICT: PASS\n- ok",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"ok",
		"LAUNCH: local\nAGREED",
		planB,
		"VERDICT: PASS\n- ok",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "done"},
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 1, FixRoundsMax: 3, Worker: "local"},
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
		t.Fatalf("expected worker to start when plans are equal after normalisation, got %d starts", fakeW.Starts)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "Dual AGREED.") {
		t.Fatalf("expected 'Dual AGREED.' in output: %s", outStr)
	}
}
