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
	q := strings.ToLower(strings.TrimSpace(ans))
	if q == "quit" || q == "exit" {
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}

	msg := fmt.Sprintf("Sahil replied:\n%s\n\nPrompt: Interpret Sahil's reply. End your response with one line:\nSAHIL_MEANS: PASS|FAIL|FIX", ans)
	ra, rb, err := sendBoth(ctx, s.A, s.B, msg, msg)
	if err != nil {
		return err
	}
	s.say(s.A, ra)
	s.say(s.B, rb)

	ma, mb := parseSahilMeans(ra), parseSahilMeans(rb)
	if ma == "PASS" && mb == "PASS" {
		return s.bothPass(run)
	}
	if ma == "FAIL" && mb == "FAIL" {
		return s.startFix(ctx, w, job, va, vb)
	}
	return s.startFix(ctx, w, job, ans, ans)
}

func parseSahilMeans(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		upper := strings.ToUpper(strings.TrimSpace(lines[i]))
		if strings.HasPrefix(upper, "SAHIL_MEANS:") {
			val := strings.TrimSpace(upper[len("SAHIL_MEANS:"):])
			switch val {
			case "PASS", "FAIL", "FIX":
				return val
			}
		}
	}
	return ""
}
