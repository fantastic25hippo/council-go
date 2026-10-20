package meeting

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/extract"
	"github.com/fantastic25hippo/council-go/internal/gate"
	"github.com/fantastic25hippo/council-go/internal/gitstat"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
)

func (s *Session) tryAgree(ctx context.Context, first member.Member, text string, second member.Member) (done bool, secondText string, err error) {
	if !extract.IsAgreed(text) {
		return false, "", nil
	}
	p, ok := extract.ParsePlan(text)
	if !ok {
		nack, err := first.Send(prompt.NackPlan(s.workerIDs()))
		if err != nil {
			return false, "", err
		}
		s.say(first, nack)
		if !extract.IsAgreed(nack) {
			return false, "", nil
		}
		p, ok = extract.ParsePlan(nack)
		if !ok {
			return false, "", nil
		}
		text = nack
	}
	s.Plan = p
	s.PlanText = text
	s.Council.Phase = state.PhaseWaitAgreed
	conf, err := second.Send(prompt.Confirm(first.DisplayName(), text, p.Worker))
	if err != nil {
		return false, "", err
	}
	s.say(second, conf)
	secondText = conf
	if !extract.IsAgreed(conf) {
		return false, conf, nil
	}
	launch := extract.ParseLaunch(conf)
	if launch != p.Worker {
		ln, err := second.Send(prompt.LaunchNack(p.Worker))
		if err != nil {
			return false, secondText, err
		}
		s.say(second, ln)
		secondText = ln
		if !extract.IsAgreed(ln) || extract.ParseLaunch(ln) != p.Worker {
			return false, ln, nil
		}
	}

	dirty, unpushed := false, false
	if s.Cfg.CheckUnpushed {
		var stErr error
		dirty, unpushed, stErr = gitstat.State(s.Cwd)
		if stErr != nil {
			s.sys("git state: " + stErr.Error())
			s.Council.Phase = state.PhaseTaskDone
			return true, secondText, nil
		}
	}
	_, err = Prepare(s.Plan, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
	if err != nil {
		if errors.Is(err, gate.ErrUnpushed) || errors.Is(err, gate.ErrDirty) {
			s.sys("Commit/push first: " + err.Error())
			s.Council.Phase = state.PhaseTaskDone
			return true, secondText, nil
		}
		s.sys("Plan NACK: " + err.Error())

		nackMsg := prompt.NackPlanReason(s.workerIDs(), err.Error())
		nackA, errA := s.A.Send(nackMsg)
		if errA != nil {
			return false, secondText, errA
		}
		s.say(s.A, nackA)

		nackB, errB := s.B.Send(nackMsg)
		if errB != nil {
			return false, secondText, errB
		}
		s.say(s.B, nackB)

		if extract.IsAgreed(nackA) && extract.IsAgreed(nackB) {
			pA, okA := extract.ParsePlan(nackA)
			pB, okB := extract.ParsePlan(nackB)
			if okA && okB {
				_, pErrA := Prepare(pA, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
				if pErrA != nil {
					if errors.Is(pErrA, gate.ErrUnpushed) || errors.Is(pErrA, gate.ErrDirty) {
						s.sys("Commit/push first: " + pErrA.Error())
						s.Council.Phase = state.PhaseTaskDone
						return true, secondText, nil
					}
					s.sys(pErrA.Error())
				}
				_, pErrB := Prepare(pB, s.ResolvedWorker, capsFor(s.ResolvedWorker), dirty, unpushed)
				if pErrB != nil {
					if errors.Is(pErrB, gate.ErrUnpushed) || errors.Is(pErrB, gate.ErrDirty) {
						s.sys("Commit/push first: " + pErrB.Error())
						s.Council.Phase = state.PhaseTaskDone
						return true, secondText, nil
					}
					s.sys(pErrB.Error())
				}
				if pErrA == nil && pErrB == nil && plansEqual(pA, pB) {
					s.Plan = pA
					s.PlanText = nackA
				s.sys("Dual AGREED.")
				return true, secondText, nil
				}
			}
		}

		s.sys("Members still failed to provide a valid plan.")
		return false, secondText, nil
	}

	s.sys("Dual AGREED.")
	return true, secondText, nil
}

func plansEqual(a, b plan.Block) bool {
	if !strings.EqualFold(normalize(a.Worker), normalize(b.Worker)) {
		return false
	}
	if normalize(a.Change) != normalize(b.Change) {
		return false
	}
	if normalize(a.Verify) != normalize(b.Verify) {
		return false
	}
	return equalNormalizedPaths(a.OwnedPaths, b.OwnedPaths)
}

func equalNormalizedPaths(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	na := make([]string, len(a))
	for i, p := range a {
		na[i] = normalize(p)
	}
	nb := make([]string, len(b))
	for i, p := range b {
		nb[i] = normalize(p)
	}
	sort.Strings(na)
	sort.Strings(nb)
	for i := range na {
		if na[i] != nb[i] {
			return false
		}
	}
	return true
}
