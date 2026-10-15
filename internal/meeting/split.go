package meeting

import (
	"context"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func (s *Session) bothPass(run worker.Run) error {
	s.sys("Both PASS.")
	if run.Raw.PRURL != "" {
		s.sys("PR: " + run.Raw.PRURL)
		s.sys("Ready for Sahil's review.")
	}
	return nil
}

func (s *Session) startFix(ctx context.Context, w worker.Worker, job worker.Job, va, vb string) error {
	if s.Council.FixRounds >= s.Cfg.FixRoundsMax {
		s.sys("Fix round cap reached. Adjourning.")
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
	s.Council.SetFixing()
	s.Council.Phase = state.PhaseFixing
	s.sys("Starting a fix round.")
	fix := job
	fix.Text = job.Text + "\n\nFIX:\n" + va + "\n" + vb
	if w.Capabilities().KeepsMemoryAcrossFixes {
		return s.handleRun(ctx, w, fix, s.runWorker(ctx, func() worker.Run { return w.Continue(job, fix.Text) }))
	}
	return s.handleRun(ctx, w, fix, s.runWorker(ctx, func() worker.Run { return w.Start(fix) }))
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
	return s.splitAskSahil(ctx, w, job)
}

func (s *Session) splitAskSahil(ctx context.Context, w worker.Worker, job worker.Job) error {
	s.sys("Split verdicts. Type a ruling, or quit.")
	ans, err := s.readSahil(ctx, "Sahil> ")
	if err != nil {
		return err
	}
	if q := strings.ToLower(strings.TrimSpace(ans)); q == "quit" || q == "exit" {
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
	return s.startFix(ctx, w, job, ans, ans)
}
