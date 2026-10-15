package meeting

import (
	"context"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func (s *Session) handleFail(ctx context.Context, w worker.Worker, job worker.Job, run worker.Run) error {
	s.sys("Worker failed.")
	if s.Council.FixRounds >= s.Cfg.FixRoundsMax {
		s.sys("Fix round cap reached.")
		s.Council.Phase = state.PhaseTaskDone
		return nil
	}
	failMsg := prompt.WorkerFailed(run.Report)
	ra, rb, err := sendBoth(ctx, s.A, s.B, failMsg, failMsg)
	if err != nil {
		return err
	}

	aFix := lastLine(ra) == "FIX"
	bFix := lastLine(rb) == "FIX"

	if aFix && bFix {
		return s.runFix(ctx, w, job, ra+"\n"+rb)
	}

	s.say(s.A, ra)
	s.say(s.B, rb)
	s.sys("Type a ruling, or quit.")
	ans, err := s.readSahil(ctx, "Sahil> ")
	if err != nil {
		return err
	}
	if q := strings.ToLower(strings.TrimSpace(ans)); q == "quit" || q == "exit" {
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}

	return s.runFix(ctx, w, job, ans)
}

func (s *Session) runFix(ctx context.Context, w worker.Worker, job worker.Job, reasons string) error {
	if s.Council.FixRounds >= s.Cfg.FixRoundsMax {
		s.sys("Fix round cap reached.")
		s.Council.Phase = state.PhaseTaskDone
		return nil
	}
	s.Council.SetFixing()
	s.Council.Phase = state.PhaseFixing
	s.sys("Starting a fix round.")
	msg := fmt.Sprintf("FIX ROUND %d:\n%s", s.Council.FixRounds, reasons)
	if w.Capabilities().KeepsMemoryAcrossFixes {
		return s.handleRun(ctx, w, job, s.runWorker(ctx, func() worker.Run { return w.Continue(job, msg) }))
	}
	fix := job
	fix.Text = job.Text + "\n\n" + msg
	return s.handleRun(ctx, w, job, s.runWorker(ctx, func() worker.Run { return w.Start(fix) }))
}

func lastLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if t != "" {
			return strings.ToUpper(t)
		}
	}
	return ""
}
