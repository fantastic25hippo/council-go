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

func (s *Session) review(ctx context.Context, w worker.Worker, job worker.Job, run worker.Run) error {
	s.Council.Phase = state.PhaseVerify
	ev := s.collect(run)
	ask := prompt.ReviewAsk(ev.Diff, ev.VerifyOutput, ev.VerifyExit, run.Report)
	va, vb, err := sendBoth(ctx, s.A, s.B, ask, ask)
	if err != nil {
		return err
	}
	s.say(s.A, va)
	s.say(s.B, vb)
	pa, pb := extract.ParseVerdict(va), extract.ParseVerdict(vb)
	s.sys(fmt.Sprintf("Verdicts: %s=%s %s=%s", s.A.DisplayName(), nz(pa), s.B.DisplayName(), nz(pb)))
	if pa == "PASS" && pb == "PASS" {
		return s.bothPass(run)
	}
	if pa == "FAIL" && pb == "FAIL" {
		return s.startFix(ctx, w, job, va, vb)
	}
	return s.splitOnce(ctx, w, job, run, va, vb)
}

func nz(s string) string {
	if strings.TrimSpace(s) == "" {
		return "?"
	}
	return s
}
