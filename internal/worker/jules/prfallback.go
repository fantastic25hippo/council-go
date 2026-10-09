package jules

import (
	"encoding/json"
)

type GHRunner interface {
	Run(argv []string) (stdout, stderr string, exit int, err error)
}

func (l *Live) fallbackPR() {
	if l.Runner == nil || l.Repo == "" {
		return
	}

	stdout, _, exit, err := l.Runner.Run([]string{"gh", "pr", "list", "--repo", l.Repo, "--state", "open", "--json", "url,title", "--limit", "5"})
	if err != nil || exit != 0 {
		return
	}

	var prs []struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(stdout), &prs); err != nil {
		return
	}

	if len(prs) == 1 && prs[0].URL != "" {
		l.sess.SetPRURL(prs[0].URL)
	}
}
