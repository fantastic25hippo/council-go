package meeting

import (
	"context"
	"os/exec"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/dispatch"
	"github.com/fantastic25hippo/council-go/internal/evidence"
	"github.com/fantastic25hippo/council-go/internal/extract"
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
	return worker.Capabilities{CanAskAndWait: true, TouchesLocalCheckout: true}
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
	id, err := Prepare(s.Task, s.Plan, s.Cfg.Worker, originGitHub(s.Cwd), capsFor(s.Plan.Worker), dirty, unpushed)
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
	act := s.Relay.OnRun(run)
	switch act {
	case dispatch.ActionAskSahil:
		if s.Council.FixRounds >= s.Cfg.FixRoundsMax {
			s.sys("Fix round cap reached. Adjourning.")
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		s.Council.SetFixing() // Count against cap
		s.Council.Phase = state.PhaseWorkerAsked
		s.sys("Worker asked: " + run.Question)

		pq := prompt.WorkerQuestion(run.Question)
		ra, rb, err := sendBoth(ctx, s.A, s.B, pq, pq)
		if err != nil {
			return err
		}

		aa, oka := extract.ParseAnswer(ra)
		ab, okb := extract.ParseAnswer(rb)

		if oka && okb && aa != "" && aa == ab {
			return s.handleRun(ctx, w, job, w.Continue(job, aa))
		}

		s.say(s.A, ra)
		s.say(s.B, rb)

		ans, err := s.readSahil(ctx, "Sahil> ")
		if err != nil {
			return err
		}
		if q := strings.ToLower(strings.TrimSpace(ans)); q == "quit" || q == "exit" {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		return s.handleRun(ctx, w, job, w.Continue(job, ans))
	case dispatch.ActionFix:
		s.sys("Worker failed.")
		return s.review(ctx, w, job, run)
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
