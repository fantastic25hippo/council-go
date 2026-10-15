package meeting_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/meeting"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/pause"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestAskedMembersAgree(t *testing.T) {
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
		"ANSWER: master",
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
	if !strings.Contains(out.String(), "Worker asked: which branch?") {
		t.Fatalf("missing worker asked: %s", out.String())
	}
	if fakeW.Continues != 1 {
		t.Fatalf("continues %d", fakeW.Continues)
	}
}

func TestAskedMembersDisagree(t *testing.T) {
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
		"ANSWER: main",
		"VERDICT: PASS\n- ok",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeAsked, Question: "which branch?"},
	}
	fakeW.HandleContinue = func(job worker.Job, ans string) worker.Run {
		if strings.TrimPrefix(ans, "ANSWER: ") != "use theirs" {
			t.Errorf("expected use theirs, got %s", ans)
		}
		return worker.Run{Outcome: worker.OutcomeDone, Report: "did it"}
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\ntask\n\nuse theirs\nquit\n"), Out: &out},
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
	if !strings.Contains(out.String(), "ANSWER: main") {
		t.Fatalf("should print both answers on split: %s", out.String())
	}
}

func TestInterject(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		"heard you",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"heard you",
	}}
	pc := pause.New()
	var out bytes.Buffer
	council := &state.Council{}
	ch := make(chan string, 10)
	ch <- "hi"
	go func() {
		time.Sleep(80 * time.Millisecond)
		pc.OnSIGINT()
		time.Sleep(80 * time.Millisecond)
		ch <- "hello both"
		time.Sleep(40 * time.Millisecond)
		ch <- "c"
		time.Sleep(40 * time.Millisecond)
		ch <- "quit"
		close(ch)
	}()
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: chanReader(ch), Out: &out, Pause: pc},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: &worker.Fake{}},
		WorkerFor: func(string) worker.Worker {
			return &worker.Fake{Next: worker.Run{Outcome: worker.OutcomeDone}}
		},
		Cwd: t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil && err != context.Canceled {
		t.Fatal(err)
	}
	found := false
	for _, m := range a.Sent {
		if strings.Contains(m, "Message from Sahil:\nhello both") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("interject not found in a's history: %v out=%s", a.Sent, out.String())
	}
}

type chanReader chan string

func (c chanReader) Read(p []byte) (n int, err error) {
	s, ok := <-c
	if !ok {
		return 0, context.Canceled
	}
	n = copy(p, s+"\n")
	return n, nil
}
