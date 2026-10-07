package evidence

import (
	"errors"
	"fmt"
)

// GHRunner runs a gh/git argv and returns stdout, stderr, exit, err.
type GHRunner interface {
	Run(argv []string) (stdout, stderr string, exit int, err error)
}

// PRDiff asks the runner for: gh pr diff <pr>
// argv must be exactly []string{"gh", "pr", "diff", pr} when pr is a number or URL.
func PRDiff(run GHRunner, pr string) (string, error) {
	if pr == "" {
		return "", errors.New("empty PR")
	}
	stdout, _, exit, err := run.Run([]string{"gh", "pr", "diff", pr})
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("gh exit %d", exit)
	}
	return stdout, nil
}

// PRChecks asks the runner for: gh pr checks <pr>
func PRChecks(run GHRunner, pr string) (string, error) {
	if pr == "" {
		return "", errors.New("empty PR")
	}
	stdout, _, exit, err := run.Run([]string{"gh", "pr", "checks", pr})
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("gh exit %d", exit)
	}
	return stdout, nil
}
