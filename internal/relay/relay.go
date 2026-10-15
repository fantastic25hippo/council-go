package relay

import (
	"github.com/fantastic25hippo/council-go/internal/dispatch"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/route"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

type Relay struct {
	Council *state.Council
	Worker  worker.Worker
}

func (r *Relay) OnAgreed(p plan.Block, resolved string, caps worker.Capabilities) error {
	return route.Validate(p, resolved, caps)
}

func (r *Relay) OnRun(run worker.Run) dispatch.Action {
	r.Council.SetRun(run)
	act := dispatch.Handle(run)

	switch act {
	case dispatch.ActionAskSahil:
		r.Council.SetAsked()
	case dispatch.ActionFix:
		r.Council.Phase = state.PhaseFixing
	case dispatch.ActionCleanup:
		r.Worker.Cancel()
		r.Council.ClearRun()
	case dispatch.ActionCollectEvidence:
		// do not Cancel
	}

	return act
}

func (r *Relay) OnSIGHUP() error {
	err := r.Worker.Cancel()
	r.Council.ClearRun()
	return err
}
