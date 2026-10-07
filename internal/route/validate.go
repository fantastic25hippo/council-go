package route

import (
	"errors"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// Validate checks if the plan block is valid for the resolved worker and capabilities.
func Validate(p plan.Block, resolved string, caps worker.Capabilities) error {
	if p.Worker != resolved {
		return errors.New("invalid WORKER in plan: NACK")
	}

	if (caps.ReachesGitHub || caps.OpensPR) && strings.TrimSpace(p.Repo) == "" {
		return errors.New("empty REPO in plan: NACK")
	}

	return nil
}
