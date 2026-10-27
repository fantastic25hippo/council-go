package meeting

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestParseSahilMeans(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"SAHIL_MEANS: PASS", "PASS"},
		{"SAHIL_MEANS: FAIL", "FAIL"},
		{"SAHIL_MEANS: FIX", "FIX"},
		{"looks good\nSAHIL_MEANS: pass", "PASS"},
		{"needs rework\nsahil_means: fix\n", "FIX"},
		{"broken\nSAHIL_MEANS: FAIL\nextra", "FAIL"},
		{"just a comment", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := parseSahilMeans(tc.input); got != tc.want {
			t.Errorf("parseSahilMeans(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSplitAskSahilMeansPassAndFail(t *testing.T) {
	ctx := context.Background()
	job := worker.Job{Text: "task"}
	run := worker.Run{Outcome: worker.OutcomeDone}

	// Dual PASS -> bothPass
	var out bytes.Buffer
	a := &member.Fake{Name: "a", Shown: "A", Replies: []string{"SAHIL_MEANS: PASS"}}
	b := &member.Fake{Name: "b", Shown: "B", Replies: []string{"SAHIL_MEANS: PASS"}}
	w := &worker.Fake{}
	s := &Session{
		A: a, B: b, Council: &state.Council{},
		IO: &term.IO{In: strings.NewReader("ship it\n"), Out: &out},
	}
	if err := s.splitAskSahil(ctx, w, job, run, "va", "vb"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Both PASS") {
		t.Fatalf("expected Both PASS, got %s", out.String())
	}

	// Dual FAIL -> startFix with member reasons
	out.Reset()
	a = &member.Fake{Name: "a", Shown: "A", Replies: []string{"SAHIL_MEANS: FAIL"}}
	b = &member.Fake{Name: "b", Shown: "B", Replies: []string{"SAHIL_MEANS: FAIL"}}
	w = &worker.Fake{
		Caps: worker.Capabilities{KeepsMemoryAcrossFixes: true},
		HandleContinue: func(_ worker.Job, fix string) worker.Run {
			return worker.Run{Outcome: worker.OutcomeUnspecified, Report: "stop"}
		},
	}
	c := &state.Council{}
	s = &Session{
		Cfg:     cfg.Config{FixRoundsMax: 3},
		A:       a,
		B:       b,
		Council: c,
		Relay:   &relay.Relay{Council: c, Worker: w},
		IO:      &term.IO{In: strings.NewReader("no way\n"), Out: &out},
	}
	if err := s.splitAskSahil(ctx, w, job, run, "reasonA", "reasonB"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.LastAnswer, "reasonA") || !strings.Contains(w.LastAnswer, "reasonB") {
		t.Fatalf("expected member reasons in LastAnswer, got %q", w.LastAnswer)
	}
}

func TestSplitAskSahilMeansFix(t *testing.T) {
	ctx := context.Background()
	job := worker.Job{Text: "task"}
	run := worker.Run{Outcome: worker.OutcomeDone}

	a := &member.Fake{Name: "a", Shown: "A", Replies: []string{"SAHIL_MEANS: FIX"}}
	b := &member.Fake{Name: "b", Shown: "B", Replies: []string{"SAHIL_MEANS: FIX"}}
	w := &worker.Fake{
		Caps: worker.Capabilities{KeepsMemoryAcrossFixes: true},
		HandleContinue: func(_ worker.Job, _ string) worker.Run {
			return worker.Run{Outcome: worker.OutcomeUnspecified, Report: "stop"}
		},
	}
	c := &state.Council{}
	s := &Session{
		Cfg:     cfg.Config{FixRoundsMax: 3},
		A:       a,
		B:       b,
		Council: c,
		Relay:   &relay.Relay{Council: c, Worker: w},
		IO:      &term.IO{In: strings.NewReader("change timeout to 5s\n"), Out: &bytes.Buffer{}},
	}
	if err := s.splitAskSahil(ctx, w, job, run, "va", "vb"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(w.LastAnswer, "change timeout to 5s") {
		t.Fatalf("expected sahil text in LastAnswer, got %q", w.LastAnswer)
	}
}

func TestPauseLoopResumeAndInterject(t *testing.T) {
	ctx := context.Background()

	// Empty line resumes
	s := &Session{
		IO:      &term.IO{In: strings.NewReader("\n"), Out: &bytes.Buffer{}},
		Council: &state.Council{Phase: state.PhaseDiscuss},
	}
	if err := s.pauseLoop(ctx); err != nil {
		t.Fatal(err)
	}

	// 'c' resumes
	s = &Session{
		IO:      &term.IO{In: strings.NewReader("c\n"), Out: &bytes.Buffer{}},
		Council: &state.Council{Phase: state.PhaseDiscuss},
	}
	if err := s.pauseLoop(ctx); err != nil {
		t.Fatal(err)
	}

	// Interject with state block
	a := &member.Fake{Name: "a", Shown: "A", Replies: []string{"A: status ok"}}
	b := &member.Fake{Name: "b", Shown: "B", Replies: []string{"B: status ok"}}
	c := &state.Council{Phase: state.PhaseWorkerRunning, WorkerID: "local"}
	s = &Session{
		A: a, B: b, Council: c,
		IO: &term.IO{In: strings.NewReader("status check?\nc\n"), Out: &bytes.Buffer{}},
	}
	if err := s.pauseLoop(ctx); err != nil {
		t.Fatal(err)
	}
	if len(a.Sent) == 0 || !strings.Contains(a.Sent[0], "status check?") {
		t.Fatalf("expected interject sent to A, got %v", a.Sent)
	}
	if !strings.Contains(a.Sent[0], "WorkerRunning") && !strings.Contains(a.Sent[0], "Phase: ") {
		t.Fatalf("expected phase in state block sent to A, got %v", a.Sent[0])
	}
}
