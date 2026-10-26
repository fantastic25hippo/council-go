package meeting_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
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

func setupFakeMemo(t *testing.T) string {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "memo.log")
	binDir := t.TempDir()
	bin := filepath.Join(binDir, "grok-agy-memo")
	_ = os.WriteFile(bin, []byte(fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %q\n", logPath)), 0755)
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
	return logPath
}

func readNoteLines(t *testing.T, path string) []string {
	t.Helper()
	data, _ := os.ReadFile(path)
	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return nil
	}
	var notes []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "note") {
			notes = append(notes, line)
		}
	}
	return notes
}

func runSession(t *testing.T, aRep, bRep []string, w *worker.Fake) string {
	t.Helper()
	var out bytes.Buffer
	c := &state.Council{}
	s := &meeting.Session{
		Cfg:       cfg.Config{Worker: "local"},
		A:         &member.Fake{Name: "grok", Shown: "Grok", Replies: aRep},
		B:         &member.Fake{Name: "gemini", Shown: "Gemini", Replies: bRep},
		IO:        &term.IO{In: strings.NewReader("hi\ntask\n\nquit\n"), Out: &out},
		Council:   c,
		Relay:     &relay.Relay{Council: c, Worker: w},
		WorkerFor: func(string) worker.Worker { return w },
		Cwd:       t.TempDir(),
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestDualAgreedRecordsNoteOnce(t *testing.T) {
	log := setupFakeMemo(t)
	w := &worker.Fake{Caps: worker.Capabilities{TouchesLocalCheckout: true}, Next: worker.Run{Outcome: worker.OutcomeFailed}}
	out := runSession(t, []string{"Hi.", agreedPlan()}, []string{"Hi.", "ok", "LAUNCH: local\nAGREED"}, w)
	if !strings.Contains(out, "Dual AGREED.") {
		t.Fatalf("expected Dual AGREED in %s", out)
	}
	lines := readNoteLines(t, log)
	if len(lines) != 1 || !strings.Contains(lines[0], "Council dual AGREED") {
		t.Fatalf("expected 1 'Council dual AGREED', got %v", lines)
	}
}

func TestBothPassRecordsNoteWithoutDualAgreed(t *testing.T) {
	log := setupFakeMemo(t)
	var out bytes.Buffer
	s := &meeting.Session{IO: &term.IO{In: strings.NewReader(""), Out: &out}, Council: &state.Council{}, Cwd: t.TempDir()}
	if err := s.BothPass(worker.Run{}); err != nil {
		t.Fatal(err)
	}
	lines := readNoteLines(t, log)
	if len(lines) != 1 || !strings.Contains(lines[0], "Council both PASS") {
		t.Fatalf("expected 1 'Council both PASS', got %v", lines)
	}
}

func TestDualAgreedAndBothPassCollapsesToOneNote(t *testing.T) {
	log := setupFakeMemo(t)
	w := &worker.Fake{Caps: worker.Capabilities{TouchesLocalCheckout: true, CanAskAndWait: true}, Next: worker.Run{Outcome: worker.OutcomeDone}}
	out := runSession(t, []string{"Hi.", agreedPlan(), "VERDICT: PASS\n- ok"}, []string{"Hi.", "ok", "LAUNCH: local\nAGREED", "VERDICT: PASS\n- ok"}, w)
	if !strings.Contains(out, "Both PASS.") {
		t.Fatalf("expected Both PASS in %s", out)
	}
	lines := readNoteLines(t, log)
	if len(lines) != 1 {
		t.Fatalf("expected 1 note, got %d: %v", len(lines), lines)
	}
}

func TestMissingMemoBinaryNonFatal(t *testing.T) {
	var kept []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(dir, "grok-agy-memo")); os.IsNotExist(err) {
			kept = append(kept, dir)
		}
	}
	t.Setenv("PATH", strings.Join(kept, string(filepath.ListSeparator)))
	w := &worker.Fake{Caps: worker.Capabilities{TouchesLocalCheckout: true}}
	out := runSession(t, []string{"Hi.", agreedPlan()}, []string{"Hi.", "ok", "LAUNCH: local\nAGREED"}, w)
	if !strings.Contains(out, "Dual AGREED.") || !strings.Contains(out, "grok-agy-memo not found") {
		t.Fatalf("expected Dual AGREED and missing memo warning in %s", out)
	}
}
