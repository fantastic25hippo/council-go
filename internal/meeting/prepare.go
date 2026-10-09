package meeting

import (
	"errors"

	"github.com/fantastic25hippo/council-go/internal/gate"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/route"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// Prepare composes existing checks before a worker starts.
// It does not mutate p. It does not start a worker.
func Prepare(p plan.Block, resolvedWorker string, caps worker.Capabilities, dirty, unpushed bool) (workerID string, err error) {
	if p.Worker != resolvedWorker {
		return "", errors.New("NACK: plan.WORKER does not match resolved worker")
	}

	if err := route.Validate(p, resolvedWorker, caps); err != nil {
		return "", err
	}

	if err := gate.ProtectedOverlap(p.OwnedPaths); err != nil {
		return "", err
	}

	if err := gate.GitHubReady(caps, dirty, unpushed); err != nil {
		return "", err
	}

	return resolvedWorker, nil
}
