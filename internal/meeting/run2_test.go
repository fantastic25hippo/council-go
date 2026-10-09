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
			Report: "did it",
			Raw: worker.Raw{PRURL: "https://github.com/o/r/pull/1"},
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
	if !strings.Contains(out.String(), "PR: https://github.com/o/r/pull/1") {
		t.Fatalf("no pr url %s", out.String())
	}
	if !strings.Contains(out.String(), "Ready for Sahil's review.") {
		t.Fatalf("no pr review msg %s", out.String())
	}
}
