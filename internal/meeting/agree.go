package meeting

import (
	"context"

	"github.com/fantastic25hippo/council-go/internal/extract"
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
		dirty, unpushed, _ = gitstat.State(s.Cwd)
	}
	_, err = Prepare(s.Plan, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err != nil {
		s.sys("Plan NACK: " + err.Error())

		nackMsg := prompt.NackPlan(s.workerIDs())
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
