package meeting

import (
	"context"
	"time"

	"github.com/fantastic25hippo/council-go/internal/attach"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/sitlog"
	"github.com/fantastic25hippo/council-go/internal/state"
)

// Run is the meeting: hi, greet, task, discuss, worker, review, adjourn.
func (s *Session) Run(ctx context.Context) error {
	if s.Log == nil && s.Cwd != "" {
		if lg, err := sitlog.Open(s.Cwd, time.Now()); err != nil {
			s.IO.Print("", "log: "+err.Error())
		} else {
			s.Log = lg
		}
	}
	defer func() {
		if s.Log != nil {
			s.Log.Close()
		}
		_ = s.A.Close()
		_ = s.B.Close()
		_ = s.Relay.OnSIGHUP()
	}()
	if err := s.waitHI(ctx); err != nil {
		return err
	}
	if s.Council.Phase == state.PhaseAdjourn {
		return nil
	}
	if err := s.greet(ctx); err != nil {
		return err
	}

	for {
		if err := s.waitTask(ctx); err != nil {
			return err
		}
		if s.Council.Phase == state.PhaseAdjourn {
			return nil
		}
		if s.Attach.SessionID != "" && s.Attach.PRNumber != "" {
			if err := s.workAttach(ctx); err != nil {
				return err
			}
			if s.Council.Phase == state.PhaseAdjourn {
				return nil
			}
			s.nextTask()
			continue
		}
		if err := s.discuss(ctx); err != nil {
			return err
		}
		if s.Council.Phase == state.PhaseAdjourn {
			return nil
		}
		if s.Plan.Worker == "" {
			s.sys("No agreed plan.")
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		if err := s.work(ctx); err != nil {
			return err
		}
		if s.Council.Phase == state.PhaseAdjourn {
			return nil
		}
		s.nextTask()
	}
}

func (s *Session) nextTask() {
	s.Attach = attach.Request{}
	s.Task = ""
	s.Plan = plan.Block{}
	s.PlanText = ""
	s.Council.FixRounds = 0
	s.Council.WorkerID = ""
	s.Council.ClearRun()
}
