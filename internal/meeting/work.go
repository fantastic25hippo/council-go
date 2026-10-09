package meeting

import (
	"context"
	"os/exec"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/dispatch"
	"github.com/fantastic25hippo/council-go/internal/evidence"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/verify"
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
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
	}
	id, err := Prepare(s.Plan, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err != nil {
		s.sys("Plan NACK: " + err.Error())
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
	w := s.WorkerFor(id)
	if w == nil {
		s.sys("no worker for " + id)
		s.Council.Phase = state.PhaseAdjourn
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
	}
	return s.handleRun(ctx, w, job, w.Start(job))
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
		s.Council.Phase = state.PhaseAdjourn
		return nil
	case dispatch.ActionCollectEvidence:
		return s.review(ctx, w, job, run)
	default:
		s.sys("Worker hard stop.")
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
}

func (s *Session) collect(run worker.Run) worker.Evidence {
	if run.Raw.PRURL != "" {
		gh := s.GH
		if gh == nil {
			gh = evidence.DefaultGHRunner()
		}
		diff, _ := evidence.PRDiff(gh, run.Raw.PRURL)
		out, err := evidence.PRChecks(gh, run.Raw.PRURL)
		if err != nil && out == "" {
			out = err.Error()
		}
		exit := 0
		if err != nil {
			exit = 1
		}
		if diff == "" {
			diff = run.Report
		}
		return evidence.Bundle(diff, out, exit)
	}

	diff, _ := evidence.LocalDiff(s.Cwd, s.Plan.OwnedPaths)
	secs := int(s.Cfg.WorkerTimeout.Seconds())
	if secs <= 0 {
		secs = 60
	}
	out, exit, err := verify.Run(s.Cwd, s.Plan.Verify, secs)
	if err != nil && out == "" {
		out = err.Error()
		if exit == 0 {
			exit = 1
		}
	}
	if diff == "" {
		diff = run.Report
	}
	return evidence.Bundle(diff, out, exit)
}
