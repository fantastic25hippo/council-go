package meeting

import (
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/evidence"
	"github.com/fantastic25hippo/council-go/internal/verify"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func (s *Session) collect(run worker.Run) worker.Evidence {
	workerID := ""
	if s.Council != nil {
		workerID = s.Council.WorkerID
	}
	if (workerID == "jules" || s.ResolvedWorker == "jules" || s.Attach.SessionID != "") && run.Raw.PRURL == "" {
		return evidence.Bundle("", "no PR found", 1)
	}

	if run.Raw.PRURL != "" {
		gh := s.GH
		if gh == nil {
			gh = evidence.DefaultGHRunner()
		}
		diff, err := evidence.PRDiff(gh, run.Raw.PRURL, s.Plan.Repo)
		if err != nil {
			diff = "(gh pr diff failed: " + err.Error() + ")"
		}
		chk, err := evidence.PRChecks(gh, run.Raw.PRURL, s.Plan.Repo)
		out := chk.Out
		if out == "" && err != nil {
			out = err.Error()
		}
		exit := chk.Exit
		if err != nil && exit == 0 {
			exit = 1
		}
		if chk.Pending {
			out = "checks pending\n" + out
		}
		if chk.Failed && !strings.Contains(out, "checks failed") {
			out = "checks failed\n" + out
		}
		if strings.TrimSpace(s.Plan.Verify) != "" {
			secs := int(s.Cfg.WorkerTimeout.Seconds())
			if secs <= 0 {
				secs = 60
			}
			vout, vexit, verr := verifyPRHead(s.Cwd, run.Raw.PRURL, s.Plan.Verify, secs)
			if out != "" && !strings.HasSuffix(out, "\n") {
				out += "\n"
			}
			out += "VERIFY (local on PR head)\n"
			if verr != nil {
				out += fmt.Sprintf("(skipped: %v)\n", verr)
			} else {
				out += vout
				if vexit != 0 {
					exit = vexit
				}
			}
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
