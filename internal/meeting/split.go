package meeting

import (
	"context"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/optmem"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// BothPass triggers bothPass directly (used by tests).
func (s *Session) BothPass(run worker.Run) error {
	return s.bothPass(run)
}

func (s *Session) noteOnce(line string) {
	if w := optmem.NoteOnce(s.Cwd, line); w != "" {
		s.sys(w)
	}
}

func (s *Session) bothPass(run worker.Run) error {
	s.sys("Both PASS.")
	s.noteOnce("Council both PASS")
	if run.Raw.PRURL != "" {
		s.sys("PR: " + run.Raw.PRURL)
		s.sys("Ready for Sahil's review.")
	}
	return nil
}

func (s *Session) startFix(ctx context.Context, w worker.Worker, job worker.Job, va, vb string) error {
	return s.runFix(ctx, w, job, va+"\n"+vb)
}

func (s *Session) splitOnce(ctx context.Context, w worker.Worker, job worker.Job, run worker.Run, va, vb string) error {
	nva, nvb, err := sendBoth(ctx, s.A, s.B,
		prompt.SplitReview(s.B.DisplayName(), vb),
		prompt.SplitReview(s.A.DisplayName(), va),
	)
	if err != nil {
		return err
	}
	s.say(s.A, nva)
	s.say(s.B, nvb)
	pa, pb := extract.ParseVerdict(nva), extract.ParseVerdict(nvb)
	s.sys(fmt.Sprintf("Verdicts: %s=%s %s=%s", s.A.DisplayName(), nz(pa), s.B.DisplayName(), nz(pb)))
	if pa == "PASS" && pb == "PASS" {
		return s.bothPass(run)
	}
	if pa == "FAIL" && pb == "FAIL" {
		return s.startFix(ctx, w, job, nva, nvb)
	}
	return s.splitAskSahil(ctx, w, job, run, nva, nvb)
}

func (s *Session) splitAskSahil(ctx context.Context, w worker.Worker, job worker.Job, run worker.Run, va, vb string) error {
	s.sys("Split verdicts. Type pass, fail, a fix instruction, or quit.")
	ans, err := s.readSahil(ctx, "Sahil> ")
	if err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "pass":
		return s.bothPass(run)
	case "fail":
		return s.startFix(ctx, w, job, va, vb)
	case "quit", "exit":
		s.Council.Phase = state.PhaseAdjourn
		return nil
	default:
		return s.startFix(ctx, w, job, ans, ans)
	}
}
