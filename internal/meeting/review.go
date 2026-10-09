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

	if (pa != "FAIL" || pb != "FAIL") && (pa != pb || pa == "" || pb == "") {
		nva, nvb, err := sendBoth(ctx, s.A, s.B, prompt.SplitReview(s.B.DisplayName(), vb), prompt.SplitReview(s.A.DisplayName(), va))
		if err != nil {
			return err
		}
		s.say(s.A, nva)
		s.say(s.B, nvb)
		va, vb = nva, nvb
		pa, pb = extract.ParseVerdict(va), extract.ParseVerdict(vb)
		s.sys(fmt.Sprintf("Verdicts: %s=%s %s=%s", s.A.DisplayName(), nz(pa), s.B.DisplayName(), nz(pb)))
	}

	if pa == "PASS" && pb == "PASS" {
		s.sys("Both PASS.")
		if run.Raw.PRURL != "" {
			s.sys("PR: " + run.Raw.PRURL)
			s.sys("Ready for Sahil's review.")
		}
		return nil
	}

	if pa != "FAIL" || pb != "FAIL" {
		s.sys(fmt.Sprintf("%s's reply:\n%s\n\n%s's reply:\n%s", s.A.DisplayName(), va, s.B.DisplayName(), vb))
		s.sys("Split verdicts. Type a ruling, or quit.")
		ruling, err := s.readSahil(ctx, "Sahil> ")
		if err != nil {
			return err
		}
		if q := strings.ToLower(strings.TrimSpace(ruling)); q == "quit" || q == "exit" {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		va = ruling
		vb = ruling
	}

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
		return s.handleRun(ctx, w, fix, w.Continue(job, fix.Text))
	}
	return s.handleRun(ctx, w, fix, w.Start(fix))
}

func nz(s string) string {
	if strings.TrimSpace(s) == "" {
		return "?"
	}
	return s
}
