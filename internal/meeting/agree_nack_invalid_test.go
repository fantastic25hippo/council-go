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

func TestAgreeNackInvalidPlanContinuesDebate(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		protectedPlan(),
		"INVALID PLAN AFTER NACK\nAGREED",
		"debate grok",
		"disputed grok",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"ok",
		"LAUNCH: local\nAGREED",
		"INVALID PLAN AFTER NACK TOO\nAGREED",
		"debate gemini",
		"disputed gemini",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
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

	outStr := out.String()
	if !strings.Contains(outStr, "Members still failed to provide a valid plan.") {
		t.Fatalf("expected 'Members still failed to provide a valid plan.' in output: %s", outStr)
	}
	debateFound := false
	for _, p := range a.Sent {
		if strings.Contains(p, "proposal/message:") {
			debateFound = true
			break
		}
	}
	if !debateFound {
		t.Fatalf("expected debate to continue after invalid NACK plan, but A received no debate prompt: %q", a.Sent)
	}
	if fakeW.Starts != 0 {
		t.Fatalf("expected worker not to start, got %d starts", fakeW.Starts)
	}
}
