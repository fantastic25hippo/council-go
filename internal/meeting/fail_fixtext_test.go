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

func TestFailFixText(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"FIX",
		"VERDICT: PASS\nlooks good",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"FIX",
		"VERDICT: PASS\nlooks good",
	}}

	var gotAnswer string
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{
			TouchesLocalCheckout:   true,
			CanAskAndWait:          true,
			KeepsMemoryAcrossFixes: true,
		},
		Next: worker.Run{Outcome: worker.OutcomeFailed, Report: "crash"},
		HandleContinue: func(job worker.Job, answer string) worker.Run {
			gotAnswer = answer
			return worker.Run{Outcome: worker.OutcomeDone, Report: "fixed"}
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
	}

	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fakeW.Continues != 1 {
		t.Fatalf("expected 1 continue, got %d", fakeW.Continues)
	}
	if !strings.Contains(gotAnswer, "FIX ROUND 1:") {
		t.Fatalf("expected answer to contain 'FIX ROUND 1:', got %q", gotAnswer)
	}
	if strings.Count(gotAnswer, "FIX ROUND 1:") > 1 {
		t.Fatalf("expected answer not to contain 'FIX ROUND 1:' twice, got %q", gotAnswer)
	}
}
