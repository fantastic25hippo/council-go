package meeting

import (
	"context"
	"errors"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/gate"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
)

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

	dirty, unpushed := false, false
	if s.Cfg.CheckUnpushed {
		var stErr error
		dirty, unpushed, stErr = gitstat.State(s.Cwd)
		if stErr != nil {
			s.sys("git state: " + stErr.Error())
			s.Council.Phase = state.PhaseTaskDone
			return true, nil
		}
	}
	_, err = Prepare(s.Plan, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err != nil {
		if errors.Is(err, gate.ErrUnpushed) || errors.Is(err, gate.ErrDirty) {
			s.sys("Commit/push first: " + err.Error())
			s.Council.Phase = state.PhaseTaskDone
			return true, nil
		}
		s.sys("Plan NACK: " + err.Error())

		nackMsg := prompt.NackPlanReason(s.workerIDs(), err.Error())
		nackA, errA := s.A.Send(nackMsg)
		if errA != nil {
			return false, errA
		}
		s.say(s.A, nackA)

		nackB, errB := s.B.Send(nackMsg)
		if errB != nil {
			return false, errB
		}
		s.say(s.B, nackB)

		if extract.IsAgreed(nackA) && extract.IsAgreed(nackB) {
			pA, okA := extract.ParsePlan(nackA)
			pB, okB := extract.ParsePlan(nackB)
			if okA && okB && pA.Worker == s.ResolvedWorker && pB.Worker == s.ResolvedWorker {
				_, pErr := Prepare(pA, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
				if pErr != nil {
					if errors.Is(pErr, gate.ErrUnpushed) || errors.Is(pErr, gate.ErrDirty) {
						s.sys("Commit/push first: " + pErr.Error())
					} else {
						s.sys(pErr.Error())
					}
					s.Council.Phase = state.PhaseTaskDone
					return true, nil
				}
				s.Plan = pA
				s.PlanText = nackA
				s.sys("Dual AGREED.")
				return true, nil
			}
		}

		s.sys("Members still failed to provide a valid plan.")
		s.Council.Phase = state.PhaseTaskDone
		return true, nil
	}

	s.sys("Dual AGREED.")
	return true, nil
}
