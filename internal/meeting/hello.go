package meeting

import (
	"context"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
)

func (s *Session) waitHI(ctx context.Context) error {
	s.Council.Phase = state.PhaseWaitHI
	s.sys("Type hi to start the council, or quit.")
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := s.readSahil(ctx, "Sahil> ")
		if err != nil {
			return err
		}
		low := strings.ToLower(strings.TrimSpace(line))
		if low == "quit" || low == "exit" {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		if low == "hi" {
			return nil
		}
	}
}

func (s *Session) greet(ctx context.Context) error {
	s.Council.Phase = state.PhaseWaitHello
	s.sys("Starting member sessions...")
	if err := s.A.Start(s.Cwd); err != nil {
		return fmt.Errorf("start %s: %w", s.A.ID(), err)
	}
	if err := s.B.Start(s.Cwd); err != nil {
		return fmt.Errorf("start %s: %w", s.B.ID(), err)
	}
	ids := s.workerIDs()
	pa := prompt.System(s.A.DisplayName(), s.B.DisplayName(), s.Cfg.DebateTurnsEach, s.Cfg.FixRoundsMax, ids) + "\n\n" + prompt.Greet(s.A.DisplayName(), s.B.DisplayName())
	pb := prompt.System(s.B.DisplayName(), s.A.DisplayName(), s.Cfg.DebateTurnsEach, s.Cfg.FixRoundsMax, ids) + "\n\n" + prompt.Greet(s.B.DisplayName(), s.A.DisplayName())
	ga, gb, err := sendBoth(ctx, s.A, s.B, pa, pb)
	if err != nil {
		return err
	}
	s.say(s.A, ga)
	s.say(s.B, gb)
	return nil
}

func sendBoth(ctx context.Context, a, b member.Member, pa, pb string) (ra, rb string, err error) {
	type result struct {
		who string
		txt string
		err error
	}
	ch := make(chan result, 2)
	go func() {
		t, e := a.Send(pa)
		ch <- result{who: "a", txt: t, err: e}
	}()
	go func() {
		t, e := b.Send(pb)
		ch <- result{who: "b", txt: t, err: e}
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case r := <-ch:
			if r.err != nil {
				return "", "", r.err
			}
			if r.who == "a" {
				ra = r.txt
			} else {
				rb = r.txt
			}
		}
	}
	return ra, rb, nil
}
