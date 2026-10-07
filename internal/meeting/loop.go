package meeting

import (
	"context"

	"github.com/fantastic25hippo/council-go/internal/state"
)

// Run is the meeting: hi, greet, task, discuss, worker, review, adjourn.
func (s *Session) Run(ctx context.Context) error {
	defer func() {
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
	if err := s.waitTask(ctx); err != nil {
		return err
	}
	if s.Council.Phase == state.PhaseAdjourn {
		return nil
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
	return s.work(ctx)
}
