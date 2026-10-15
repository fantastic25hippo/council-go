package meeting

import (
	"context"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func (s *Session) runWorker(ctx context.Context, f func() worker.Run) worker.Run {
	ch := make(chan worker.Run, 1)
	go func() { ch <- f() }()
	select {
	case run := <-ch:
		return run
	case <-ctx.Done():
		if s.Relay != nil && s.Relay.Worker != nil {
			_ = s.Relay.Worker.Cancel()
		}
		return <-ch
	}
}
