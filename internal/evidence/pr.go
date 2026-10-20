package evidence

import (
	"errors"
	"fmt"
	"strings"
)

// GHRunner runs a gh/git argv and returns stdout, stderr, exit, err.
type GHRunner interface {
	Run(argv []string) (stdout, stderr string, exit int, err error)
}

// Checks holds output and status from gh pr checks.
type Checks struct {
	Out     string
	Exit    int
	Pending bool
	Failed  bool
}

func parseRepo(pr string) string {
	idx := strings.Index(pr, "github.com/")
	if idx == -1 {
		return ""
	}
	s := pr[idx+len("github.com/"):]
	parts := strings.Split(s, "/")
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
		return parts[0] + "/" + strings.TrimSuffix(parts[1], ".git")
	}
	return ""
}

// PRDiff asks the runner for: gh pr diff <pr>
func PRDiff(run GHRunner, pr, repo string) (string, error) {
	if pr == "" {
		return "", errors.New("empty PR")
	}
	if repo == "" {
		repo = parseRepo(pr)
	}
	argv := []string{"gh", "pr", "diff", pr}
	if repo != "" {
		argv = append(argv, "--repo", repo)
	}
	stdout, _, exit, err := run.Run(argv)
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("gh exit %d", exit)
	}
	return stdout, nil
}

// PRChecks asks the runner for: gh pr checks <pr>
func PRChecks(run GHRunner, pr, repo string) (Checks, error) {
	if pr == "" {
		return Checks{}, errors.New("empty PR")
	}
	if repo == "" {
		repo = parseRepo(pr)
	}
	argv := []string{"gh", "pr", "checks", pr}
	if repo != "" {
		argv = append(argv, "--repo", repo)
	}
	stdout, stderr, exit, err := run.Run(argv)
	out := stdout
	if stderr != "" {
		if out != "" {
			out += "\n" + stderr
		} else {
			out = stderr
		}
	}
	if err != nil {
		return Checks{Out: out, Exit: exit}, err
	}
	failed := false
	if exit != 0 && exit != 8 {
		combined := stdout + "\n" + stderr
		noChecks := strings.Contains(strings.ToLower(combined), "no checks reported")
		failed = !noChecks
	}
	return Checks{
		Out:     out,
		Exit:    exit,
		Pending: exit == 8,
		Failed:  failed,
	}, nil
}
