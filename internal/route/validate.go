package route

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/verify"
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

	if strings.TrimSpace(p.Verify) == "" {
		return errors.New("empty VERIFY in plan: NACK")
	} else if err := verify.CheckSyntax(p.Verify); err != nil {
		return fmt.Errorf("invalid VERIFY in plan: NACK: %w", err)
	}

	return nil
}
