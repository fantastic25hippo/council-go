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

type fakeGHRunner struct {
	argvs [][]string
}

func (f *fakeGHRunner) Run(argv []string) (string, string, int, error) {
	f.argvs = append(f.argvs, argv)
	return "fake gh output", "", 0, nil
}

func TestRunAgreeWorkPassShowsPR(t *testing.T) {
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
		Next: worker.Run{
			Outcome: worker.OutcomeDone,
			Report:  "did it",
			Raw:     worker.Raw{PRURL: "https://github.com/o/r/pull/1"},
		},
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
		GH:  &fakeGHRunner{},
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "PR: https://github.com/o/r/pull/1") {
		t.Fatalf("no pr url %s", out.String())
	}
	if !strings.Contains(out.String(), "Ready for Sahil's review.") {
		t.Fatalf("no pr review msg %s", out.String())
	}

	gh := s.GH.(*fakeGHRunner)
	if len(gh.argvs) != 2 {
		t.Fatalf("expected 2 gh calls, got %d", len(gh.argvs))
	}
	if gh.argvs[0][2] != "diff" || gh.argvs[1][2] != "checks" {
		t.Fatalf("expected diff then checks, got %v", gh.argvs)
	}
}

func TestRunFailedRequiresFix(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"FIX",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"FIX",
	}}

	firstRun := true

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true, KeepsMemoryAcrossFixes: true},
	}
	fakeW.HandleContinue = func(job worker.Job, answer string) worker.Run {
		return worker.Run{Outcome: worker.OutcomeUnspecified, Report: "done"} // Unspecified mapped to HardStop
	}
	fakeW.Next = worker.Run{Outcome: worker.OutcomeFailed, Report: "crash"}

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

	// We inject logic so the first start fails, and the members say FIX, which leads to continue.
	// HardStop will end the meeting.
	_ = firstRun // used conceptually

	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fakeW.Starts != 1 {
		t.Fatalf("expected 1 start, got %d", fakeW.Starts)
	}
	if fakeW.Continues != 1 {
		t.Fatalf("expected 1 continue, got %d", fakeW.Continues)
	}

	if !strings.Contains(out.String(), "Worker failed.") {
		t.Fatalf("expected worker failed message, got %s", out.String())
	}
}
