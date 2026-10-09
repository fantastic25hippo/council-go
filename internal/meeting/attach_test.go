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

func TestRunAttachJoinsSessionAndPR(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		"VERDICT: PASS\n- looks good",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"VERDICT: PASS\n- ok",
	}}
	w := &worker.Fake{
		Caps: worker.Capabilities{CanAskAndWait: true, ReachesGitHub: true, OpensPR: true, KeepsMemoryAcrossFixes: true},
		Next: worker.Run{
			Outcome: worker.OutcomeDone,
			Report:  "already done",
			Raw:     worker.Raw{PRURL: "https://github.com/o/r/pull/8"},
		},
	}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "auto"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nhttps://jules.google.com/session/1234567890 https://github.com/o/r/pull/8\nplease review\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: w},
		WorkerFor: func(id string) worker.Worker {
			if id != "jules" {
				t.Errorf("worker id %s", id)
			}
			return w
		},
		GH:  &fakeGHRunner{},
		Cwd: t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if w.Starts != 1 {
		t.Fatalf("starts %d", w.Starts)
	}
	if w.LastJob.SessionID != "1234567890" {
		t.Fatalf("session %q", w.LastJob.SessionID)
	}
	if !strings.Contains(w.LastJob.PRURL, "/pull/8") {
		t.Fatalf("pr %q", w.LastJob.PRURL)
	}
	if !strings.Contains(out.String(), "Both PASS") {
		t.Fatalf("out %s", out.String())
	}
	for _, p := range a.Sent {
		if strings.Contains(p, "This task will be run by") {
			t.Fatalf("discuss ran: %q", p)
		}
	}
}

func TestRunAttachSessionWithoutPRReprompts(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{"Hello, I am Grok."}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{"Hello, I am Gemini."}}
	w := &worker.Fake{Caps: worker.Capabilities{OpensPR: true, ReachesGitHub: true}}
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "auto"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nhttps://jules.google.com/session/1234567890\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: w},
		WorkerFor: func(string) worker.Worker {
			return w
		},
		Cwd: t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if w.Starts != 0 {
		t.Fatalf("starts %d", w.Starts)
	}
	if !strings.Contains(out.String(), "Attach needs a PR") {
		t.Fatalf("out %s", out.String())
	}
}
