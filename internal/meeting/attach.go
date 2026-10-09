package meeting

import (
	"context"
	"fmt"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func (s *Session) workAttach(ctx context.Context) error {
	repo := s.Attach.Repo
	if repo == "" {
		repo = gitstat.OriginRepo(s.Cwd)
	}
	if repo == "" {
		s.sys("Attach needs owner/repo (paste a GitHub PR URL).")
		return nil
	}

	PRURL := fmt.Sprintf("https://github.com/%s/pull/%s", repo, s.Attach.PRNumber)

	w := s.WorkerFor("jules")
	if w == nil {
		s.sys("no worker for jules")
		return nil
	}

	job := worker.Job{
		Text:      s.Task,
		Repo:      repo,
		SessionID: s.Attach.SessionID,
		PRURL:     PRURL,
	}

	s.Relay.Worker = w
	s.Council.SetRunning("jules", state.Handle{})

	s.logLine("worker", "Attach start")

	return s.handleRun(ctx, w, job, w.Start(job))
}
