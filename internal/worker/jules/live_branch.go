package jules

import "strings"

func (l *Live) defaultBranch(repo string) string {
	if l.Runner == nil || repo == "" {
		return "master"
	}
	stdout, _, exit, err := l.Runner.Run([]string{
		"gh", "repo", "view", repo, "--json", "defaultBranchRef", "-q", ".defaultBranchRef.name",
	})
	if err != nil || exit != 0 {
		return "master"
	}
	trimmed := strings.TrimSpace(stdout)
	if trimmed != "" {
		return trimmed
	}
	return "master"
}
