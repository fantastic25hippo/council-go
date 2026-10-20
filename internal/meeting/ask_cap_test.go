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

func TestAskedNormalizeCaseAndSpaces(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"ANSWER: master",
		"VERDICT: PASS\n- looks good",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"ANSWER:   Master  ",
		"VERDICT: PASS\n- ok",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeAsked, Question: "which branch?"},
	}
	fakeW.HandleContinue = func(job worker.Job, ans string) worker.Run {
		if strings.TrimPrefix(ans, "ANSWER: ") != "master" {
			t.Errorf("expected master, got %s", ans)
		}
		return worker.Run{Outcome: worker.OutcomeDone, Report: "did it"}
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\ntask\n\nquit\n"), Out: &out},
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
	if council.FixRounds != 0 {
		t.Fatalf("expected FixRounds==0, got %d", council.FixRounds)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "ANSWER: master") || !strings.Contains(outStr, "ANSWER:   Master  ") {
		t.Fatalf("output should show both member replies: %s", outStr)
	}
}

func TestAskedCapReached(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"ANSWER: master",
		"ANSWER: master",
		"ANSWER: master",
		"ANSWER: master",
		"ANSWER: master",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"ANSWER: Master",
		"ANSWER: Master",
		"ANSWER: Master",
		"ANSWER: Master",
		"ANSWER: Master",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeAsked, Question: "which branch?"},
	}
	fakeW.HandleContinue = func(job worker.Job, ans string) worker.Run {
		return worker.Run{Outcome: worker.OutcomeAsked, Question: "which branch again?"}
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\ntask\n\nquit\n"), Out: &out},
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
	if council.FixRounds != 0 {
		t.Fatalf("expected FixRounds==0, got %d", council.FixRounds)
	}
	if fakeW.Continues != 5 {
		t.Fatalf("expected 5 continues, got %d", fakeW.Continues)
	}
	if fakeW.Cancels != 2 {
		t.Fatalf("expected 2 cancels, got %d", fakeW.Cancels)
	}
	if !council.Handle.Zero() {
		t.Fatalf("expected handle to be zero after ClearRun")
	}
	if !strings.Contains(out.String(), "Ask cap reached.") {
		t.Fatalf("missing 'Ask cap reached.' in output: %s", out.String())
	}
}
