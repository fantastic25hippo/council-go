package main

import (
	"context"

	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func newRelay(w worker.Worker) *relay.Relay {
	return &relay.Relay{
		Council: &state.Council{},
		Worker:  w,
	}
}

func run(ctx context.Context, r *relay.Relay) error {
	<-ctx.Done()
	return r.OnSIGHUP()
}
