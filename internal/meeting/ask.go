package meeting

import (
	"context"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func normalize(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func (s *Session) handleAsk(ctx context.Context, w worker.Worker, job worker.Job, run worker.Run) error {
	if s.Council.Asks >= 5 {
		s.sys("Ask cap reached.")
		s.Council.Phase = state.PhaseTaskDone
		return nil
	}
	s.Council.Asks++
	s.Council.Phase = state.PhaseWorkerAsked
	s.sys("Worker asked: " + run.Question)

	pq := prompt.WorkerQuestion(run.Question)
	ra, rb, err := sendBoth(ctx, s.A, s.B, pq, pq)
	if err != nil {
		return err
	}

	aa, oka := extract.ParseAnswer(ra)
	ab, okb := extract.ParseAnswer(rb)

	s.say(s.A, ra)
	s.say(s.B, rb)

	if oka && okb && aa != "" && strings.EqualFold(normalize(aa), normalize(ab)) {
		return s.handleRun(ctx, w, job, s.runWorker(ctx, func() worker.Run { return w.Continue(job, aa) }))
	}

	ans, err := s.readSahil(ctx, "Sahil> ")
	if err != nil {
		return err
	}
	if q := strings.ToLower(strings.TrimSpace(ans)); q == "quit" || q == "exit" {
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
	return s.handleRun(ctx, w, job, s.runWorker(ctx, func() worker.Run { return w.Continue(job, ans) }))
}
