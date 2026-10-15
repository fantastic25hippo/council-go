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

func TestConfirmNACK(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"debate",
		"disputed",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"I prefer jules",
		"OBJECTION-XYZ: local cannot reach GitHub",
		"debate",
		"disputed",
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

	found := false
	for _, p := range a.Sent {
		if strings.Contains(p, "OBJECTION-XYZ") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected some prompt in a.Sent to contain %q, got: %q", "OBJECTION-XYZ", a.Sent)
	}
}

func TestLaunchNACK(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"debate",
		"disputed",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"I prefer jules",
		"AGREED",
		"OBJECTION-LAUNCH: still missing launch",
		"debate",
		"disputed",
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

	found := false
	for _, p := range a.Sent {
		if strings.Contains(p, "OBJECTION-LAUNCH") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected some prompt in a.Sent to contain %q, got: %q", "OBJECTION-LAUNCH", a.Sent)
	}
}
