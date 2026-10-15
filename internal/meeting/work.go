package meeting

import (
	"context"
	"os/exec"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/dispatch"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func capsFor(id string) worker.Capabilities {
	if id == "jules" {
		return worker.Capabilities{CanAskAndWait: true, KeepsMemoryAcrossFixes: true, ReachesGitHub: true, OpensPR: true}
	}
	return worker.Capabilities{CanAskAndWait: true, KeepsMemoryAcrossFixes: true, TouchesLocalCheckout: true}
}

func originGitHub(dir string) bool {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return false
	}
	u := strings.ToLower(string(out))
	return strings.Contains(u, "github.com")
}

func jobText(id string, p plan.Block, caps worker.Capabilities) string {
	if id == "jules" {
		return prompt.JulesJob(p, caps)
	}
	return prompt.LocalJob(p, caps)
}

func (s *Session) work(ctx context.Context) error {
	s.Council.Phase = state.PhaseWorkerRunning
	dirty, unpushed := false, false
	var err error
	if s.Cfg.CheckUnpushed {
		dirty, unpushed, err = gitstat.State(s.Cwd)
		if err != nil {
			s.sys("git state: " + err.Error())
			s.Council.Phase = state.PhaseTaskDone
			return nil
		}
	}
	id, err := Prepare(s.Plan, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err != nil {
		s.sys("Plan NACK: " + err.Error())
		s.Council.Phase = state.PhaseTaskDone
		return nil
	}
	w := s.WorkerFor(id)
	if w == nil {
		s.sys("no worker for " + id)
		s.Council.Phase = state.PhaseTaskDone
		return nil
	}
	s.Relay.Worker = w
	s.Council.SetRunning(id, state.Handle{})
	job := worker.Job{
		Plan:   s.PlanText,
		Text:   jobText(id, s.Plan, w.Capabilities()),
		Paths:  s.Plan.OwnedPaths,
		Verify: s.Plan.Verify,
		Repo:   s.Plan.Repo,
		Branch: gitstat.HeadBranch(s.Cwd),
	}
	return s.handleRun(ctx, w, job, s.runWorker(ctx, func() worker.Run { return w.Start(job) }))
}

func (s *Session) handleRun(ctx context.Context, w worker.Worker, job worker.Job, run worker.Run) error {
	s.logLine("worker", run.Outcome.String()+" "+run.Report)
	act := s.Relay.OnRun(run)
	switch act {
	case dispatch.ActionAskSahil:
		return s.handleAsk(ctx, w, job, run)
	case dispatch.ActionFix:
		return s.handleFail(ctx, w, job, run)
	case dispatch.ActionCleanup:
		s.sys("Worker cancelled.")
		if ctx.Err() != nil {
			s.Council.Phase = state.PhaseAdjourn
		} else {
			s.Council.Phase = state.PhaseTaskDone
		}
		return nil
	case dispatch.ActionCollectEvidence:
		return s.review(ctx, w, job, run)
	default:
		s.sys("Worker hard stop.")
		s.Council.Phase = state.PhaseTaskDone
		return nil
	}
}
