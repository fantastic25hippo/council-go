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

func TestFailFixRoundCapReached(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"FIX",
		"FIX",
		"FIX",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"FIX",
		"FIX",
		"FIX",
	}}

	fakeW := &worker.Fake{
		Caps: worker.Capabilities{
			TouchesLocalCheckout:   true,
			CanAskAndWait:          true,
			KeepsMemoryAcrossFixes: true,
		},
		Next: worker.Run{Outcome: worker.OutcomeFailed, Report: "crash"},
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
	if fakeW.Continues != 3 {
		t.Fatalf("expected 3 continues, got %d", fakeW.Continues)
	}
	if !strings.Contains(out.String(), "Fix round cap reached") {
		t.Fatalf("expected output to contain 'Fix round cap reached', got %s", out.String())
	}
}
