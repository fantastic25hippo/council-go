package meeting

import (
	"context"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
)

func (s *Session) waitTask(ctx context.Context) error {
	s.sys("Council is ready. Paste a task, then an empty line, or quit.")
	var lines []string
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := s.readSahil(ctx, "Sahil> ")
		if err != nil {
			return err
		}
		low := strings.ToLower(strings.TrimSpace(line))
		if len(lines) == 0 && (low == "quit" || low == "exit") {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		if strings.TrimSpace(line) == "" {
			if len(lines) == 0 {
				continue
			}
			s.Task = strings.Join(lines, "\n")
			return nil
		}
		lines = append(lines, line)
	}
}

func (s *Session) discuss(ctx context.Context) error {
	s.Council.Phase = state.PhaseDiscuss
	ids := s.workerIDs()
	pa := prompt.Proposal(s.Task, ids)
	pb := prompt.Proposal(s.Task, ids)
	ra, rb, err := sendBoth(ctx, s.A, s.B, pa, pb)
	if err != nil {
		return err
	}
	s.say(s.A, ra)
	s.say(s.B, rb)
	latestA, latestB := ra, rb
	if done, err := s.tryAgree(ctx, s.A, ra, s.B); err != nil || done {
		return err
	}
	if done, err := s.tryAgree(ctx, s.B, rb, s.A); err != nil || done {
		return err
	}
	cap := s.Cfg.DebateTurnsEach
	if cap <= 0 {
		cap = 3
	}
	for round := 0; round < cap; round++ {
		ta, err := s.A.Send(prompt.Debate(s.B.DisplayName(), latestB))
		if err != nil {
			return err
		}
		s.say(s.A, ta)
		latestA = ta
		if done, err := s.tryAgree(ctx, s.A, ta, s.B); err != nil || done {
			return err
		}
		tb, err := s.B.Send(prompt.Debate(s.A.DisplayName(), latestA))
		if err != nil {
			return err
		}
		s.say(s.B, tb)
		latestB = tb
		if done, err := s.tryAgree(ctx, s.B, tb, s.A); err != nil || done {
			return err
		}
	}
	s.sys(fmt.Sprintf("Debate cap reached. Disputed points go to Sahil."))
	da, err := s.A.Send(prompt.Disputed(cap))
	if err != nil {
		return err
	}
	db, err := s.B.Send(prompt.Disputed(cap))
	if err != nil {
		return err
	}
	s.say(s.A, da)
	s.say(s.B, db)
	s.sys("Type a ruling, or quit.")
	ruling, err := s.readSahil(ctx, "Sahil> ")
	if err != nil {
		return err
	}
	if q := strings.ToLower(strings.TrimSpace(ruling)); q == "quit" || q == "exit" {
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
	ra, rb, err = sendBoth(ctx, s.A, s.B, prompt.Ruling(ruling), prompt.Ruling(ruling))
	if err != nil {
		return err
	}
	s.say(s.A, ra)
	s.say(s.B, rb)
	if done, err := s.tryAgree(ctx, s.A, ra, s.B); err != nil || done {
		return err
	}
	if done, err := s.tryAgree(ctx, s.B, rb, s.A); err != nil || done {
		return err
	}
	s.sys("Still no dual AGREED. Adjourning.")
	s.Council.Phase = state.PhaseAdjourn
	return nil
}

func (s *Session) tryAgree(ctx context.Context, first member.Member, text string, second member.Member) (bool, error) {
	if !extract.IsAgreed(text) {
		return false, nil
	}
	p, ok := extract.ParsePlan(text)
	if !ok {
		nack, err := first.Send(prompt.NackPlan(s.workerIDs()))
		if err != nil {
			return false, err
		}
		s.say(first, nack)
		if !extract.IsAgreed(nack) {
			return false, nil
		}
		p, ok = extract.ParsePlan(nack)
		if !ok {
			return false, nil
		}
		text = nack
	}
	s.Plan = p
	s.PlanText = text
	s.Council.Phase = state.PhaseWaitAgreed
	conf, err := second.Send(prompt.Confirm(first.DisplayName(), text, p.Worker))
	if err != nil {
		return false, err
	}
	s.say(second, conf)
	if !extract.IsAgreed(conf) {
		return false, nil
	}
	launch := extract.ParseLaunch(conf)
	if launch != p.Worker {
		ln, err := second.Send(prompt.LaunchNack(p.Worker))
		if err != nil {
			return false, err
		}
		s.say(second, ln)
		if !extract.IsAgreed(ln) || extract.ParseLaunch(ln) != p.Worker {
			return false, nil
		}
	}
	s.sys("Dual AGREED.")
	return true, nil
}
