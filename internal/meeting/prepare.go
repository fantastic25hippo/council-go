package meeting

import (
	"github.com/fantastic25hippo/council-go/internal/gate"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/route"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// Prepare composes existing checks before a worker starts.
// It does not mutate p. It does not start a worker.
func Prepare(task string, p plan.Block, relayWorker string, originGitHub bool, caps worker.Capabilities, dirty, unpushed bool) (workerID string, err error) {
	override, err := plan.ParseOverride(task)
	if err != nil {
		return "", err
	}

	id, err := route.Resolve(override, relayWorker, originGitHub)
	if err != nil {
		return "", err
	}

	if err := route.Validate(p, id, caps); err != nil {
		return "", err
	}

	if err := gate.ProtectedOverlap(p.OwnedPaths); err != nil {
		return "", err
	}

	if err := gate.GitHubReady(caps, dirty, unpushed); err != nil {
		return "", err
	}

	return id, nil
}
