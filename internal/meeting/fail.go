package meeting

import (
	"context"
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
		s.Council.SetFixing()
		s.Council.Phase = state.PhaseFixing
		s.sys("Starting a fix round.")
		fix := job
		fix.Text = job.Text + "\n\nFIX:\n" + ra + "\n" + rb
		if w.Capabilities().KeepsMemoryAcrossFixes {
			return s.handleRun(ctx, w, fix, s.runWorker(ctx, func() worker.Run { return w.Continue(job, fix.Text) }))
		}
		return s.handleRun(ctx, w, fix, s.runWorker(ctx, func() worker.Run { return w.Start(fix) }))
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

	s.Council.SetFixing()
	s.Council.Phase = state.PhaseFixing
	s.sys("Starting a fix round from Sahil ruling.")
	fix := job
	fix.Text = job.Text + "\n\nFIX:\n" + ans
	if w.Capabilities().KeepsMemoryAcrossFixes {
		return s.handleRun(ctx, w, fix, s.runWorker(ctx, func() worker.Run { return w.Continue(job, fix.Text) }))
	}
	return s.handleRun(ctx, w, fix, s.runWorker(ctx, func() worker.Run { return w.Start(fix) }))
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
