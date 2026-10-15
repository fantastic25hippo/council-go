package meeting_test

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/meeting"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func julesPlan() string {
	return "WORKER: jules\nREPO: o/r\nOWNED_PATHS:\n- foo\nCHANGE: c\nVERIFY: echo hi\nAGREED"
}

func protectedPlan() string {
	return "WORKER: local\nOWNED_PATHS:\n- hook.log\nCHANGE: c\nVERIFY: echo hi\nAGREED"
}

func TestAgreeGitStateError(t *testing.T) {
	old := gitstat.Fetcher
	defer func() { gitstat.Fetcher = old }()
	gitstat.Fetcher = func(string) error { return errors.New("simulated fetch fail") }

	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{"Hi.", agreedPlan()}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{"Hi.", "ok", "LAUNCH: local\nAGREED"}}
	fakeW := &worker.Fake{Caps: worker.Capabilities{TouchesLocalCheckout: true}}
	var out bytes.Buffer
	c := &state.Council{}
	s := &meeting.Session{
		Cfg:       cfg.Config{CheckUnpushed: true, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:         a,
		B:         b,
		IO:        &term.IO{In: strings.NewReader("hi\ntask\n\nquit\n"), Out: &out},
		Council:   c,
		Relay:     &relay.Relay{Council: c, Worker: fakeW},
		WorkerFor: func(string) worker.Worker { return fakeW },
		Cwd:       t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "git state:") {
		t.Fatalf("expected 'git state:' in %s", out.String())
	}
	for _, m := range []*member.Fake{a, b} {
		for _, sent := range m.Sent {
			if strings.Contains(sent, "Resend the full plan") || strings.Contains(sent, "missing or invalid") {
				t.Fatalf("member was sent NackPlanReason text: %s", sent)
			}
		}
	}
	if c.Phase != state.PhaseAdjourn {
		t.Fatalf("phase %s", c.Phase)
	}
}

func TestAgreeGitHubReadyUnpushed(t *testing.T) {
	old := gitstat.Fetcher
	defer func() { gitstat.Fetcher = old }()
	gitstat.Fetcher = func(string) error { return nil }

	dir := t.TempDir()
	_ = exec.Command("git", "-C", dir, "init").Run()
	_ = exec.Command("git", "-C", dir, "config", "user.email", "t@e.com").Run()
	_ = exec.Command("git", "-C", dir, "config", "user.name", "t").Run()
	if err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").Run(); err != nil {
		t.Fatal(err)
	}

	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{"Hi.", julesPlan()}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{"Hi.", "ok", "LAUNCH: jules\nAGREED"}}
	fakeW := &worker.Fake{Caps: worker.Capabilities{ReachesGitHub: true}}
	var out bytes.Buffer
	c := &state.Council{}
	s := &meeting.Session{
		Cfg:            cfg.Config{CheckUnpushed: true, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "jules"},
		ResolvedWorker: "jules",
		A:              a,
		B:              b,
		IO:             &term.IO{In: strings.NewReader("hi\ntask\n\nquit\n"), Out: &out},
		Council:        c,
		Relay:          &relay.Relay{Council: c, Worker: fakeW},
		WorkerFor:      func(string) worker.Worker { return fakeW },
		Cwd:            dir,
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Commit/push first:") || !strings.Contains(out.String(), "unpushed") {
		t.Fatalf("expected 'Commit/push first:' and 'unpushed' in %s", out.String())
	}
	for _, m := range []*member.Fake{a, b} {
		for _, sent := range m.Sent {
			if strings.Contains(sent, "Resend the full plan") {
				t.Fatalf("member sent 'Resend the full plan': %s", sent)
			}
		}
	}
}

func TestAgreeProtectedOverlapNackThenAgreed(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hi.", protectedPlan(), agreedPlan(), "VERDICT: PASS\n- ok",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hi.", "ok", "LAUNCH: local\nAGREED", agreedPlan(), "VERDICT: PASS\n- ok",
	}}
	fakeW := &worker.Fake{
		Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true},
		Next: worker.Run{Outcome: worker.OutcomeDone, Report: "did it"},
	}
	var out bytes.Buffer
	c := &state.Council{}
	s := &meeting.Session{
		Cfg:       cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:         a,
		B:         b,
		IO:        &term.IO{In: strings.NewReader("hi\ntask\n\nquit\n"), Out: &out},
		Council:   c,
		Relay:     &relay.Relay{Council: c, Worker: fakeW},
		WorkerFor: func(string) worker.Worker { return fakeW },
		Cwd:       t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	foundA, foundB := false, false
	for _, sent := range a.Sent {
		if strings.Contains(sent, "protected") {
			foundA = true
		}
	}
	for _, sent := range b.Sent {
		if strings.Contains(sent, "protected") {
			foundB = true
		}
	}
	if !foundA || !foundB {
		t.Fatalf("expected members to receive nack containing 'protected', got a=%v b=%v", foundA, foundB)
	}
	if !strings.Contains(out.String(), "Dual AGREED") {
		t.Fatalf("expected 'Dual AGREED' in %s", out.String())
	}
}
