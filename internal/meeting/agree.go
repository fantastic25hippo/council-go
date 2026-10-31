package meeting

import (
	"context"
	"errors"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/gate"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
)

func (s *Session) tryAgree(ctx context.Context, first member.Member, text string, second member.Member) (bool, string, error) {
	if !extract.IsAgreed(text) {
		return false, "", nil
	}
	p, planText, ok, err := s.resolveFirstPlan(first, text)
	if err != nil || !ok {
		return false, "", err
	}
	s.Plan = p
	s.PlanText = planText
	s.Council.Phase = state.PhaseWaitAgreed

	secText, ok, err := s.confirmSecond(second, first.DisplayName(), planText, p.Worker)
	if err != nil || !ok {
		return false, secText, err
	}

	dirty, unpushed, stop := s.gitGate()
	if stop {
		return true, secText, nil
	}

	done, err := s.validateAndPrepare(dirty, unpushed)
	return done, secText, err
}

func (s *Session) resolveFirstPlan(first member.Member, text string) (plan.Block, string, bool, error) {
	p, ok := extract.ParsePlan(text)
	if ok {
		return p, text, true, nil
	}
	nack, err := first.Send(prompt.NackPlan(s.workerIDs()))
	if err != nil {
		return plan.Block{}, "", false, err
	}
	s.say(first, nack)
	if !extract.IsAgreed(nack) {
		return plan.Block{}, "", false, nil
	}
	p, ok = extract.ParsePlan(nack)
	if !ok {
		return plan.Block{}, "", false, nil
	}
	return p, nack, true, nil
}

func (s *Session) confirmSecond(second member.Member, firstName, planText, worker string) (string, bool, error) {
	conf, err := second.Send(prompt.Confirm(firstName, planText, worker))
	if err != nil {
		return "", false, err
	}
	s.say(second, conf)
	if !extract.IsAgreed(conf) {
		return conf, false, nil
	}
	if extract.ParseLaunch(conf) == worker {
		return conf, true, nil
	}
	ln, err := second.Send(prompt.LaunchNack(worker))
	if err != nil {
		return conf, false, err
	}
	s.say(second, ln)
	if !extract.IsAgreed(ln) || extract.ParseLaunch(ln) != worker {
		return ln, false, nil
	}
	return ln, true, nil
}

func (s *Session) gitGate() (dirty, unpushed, stop bool) {
	if !s.Cfg.CheckUnpushed {
		return false, false, false
	}
	dirty, unpushed, err := gitstat.State(s.Cwd)
	if err != nil {
		s.sys("git state: " + err.Error())
		s.Council.Phase = state.PhaseTaskDone
		return false, false, true
	}
	return dirty, unpushed, false
}

func (s *Session) validateAndPrepare(dirty, unpushed bool) (bool, error) {
	_, err := Prepare(s.Plan, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err == nil {
		s.sys("Dual AGREED.")
		return true, nil
	}
	if errors.Is(err, gate.ErrUnpushed) || errors.Is(err, gate.ErrDirty) {
		s.sys("Commit/push first: " + err.Error())
		s.Council.Phase = state.PhaseTaskDone
		return true, nil
	}
	return s.recoverDualNack(err, dirty, unpushed)
}

func (s *Session) recoverDualNack(reason error, dirty, unpushed bool) (bool, error) {
	s.sys("Plan NACK: " + reason.Error())
	nackMsg := prompt.NackPlanReason(s.workerIDs(), reason.Error())

	nackA, err := s.A.Send(nackMsg)
	if err != nil {
		return false, err
	}
	s.say(s.A, nackA)

	nackB, err := s.B.Send(nackMsg)
	if err != nil {
		return false, err
	}
	s.say(s.B, nackB)

	if extract.IsAgreed(nackA) && extract.IsAgreed(nackB) {
		pA, okA := extract.ParsePlan(nackA)
		pB, okB := extract.ParsePlan(nackB)
		if okA && okB {
			stopA, pErrA := s.checkPlanGate(pA, dirty, unpushed)
			if stopA {
				return true, nil
			}
			stopB, pErrB := s.checkPlanGate(pB, dirty, unpushed)
			if stopB {
				return true, nil
			}
			if pErrA == nil && pErrB == nil && plansEqual(pA, pB) {
				s.Plan = pA
				s.PlanText = nackA
				s.sys("Dual AGREED.")
				return true, nil
			}
		}
	}

	s.sys("Members still failed to provide a valid plan.")
	return false, nil
}

func (s *Session) checkPlanGate(p plan.Block, dirty, unpushed bool) (bool, error) {
	_, err := Prepare(p, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err != nil {
		if errors.Is(err, gate.ErrUnpushed) || errors.Is(err, gate.ErrDirty) {
			s.sys("Commit/push first: " + err.Error())
			s.Council.Phase = state.PhaseTaskDone
			return true, err
		}
		s.sys(err.Error())
	}
	return false, err
}
