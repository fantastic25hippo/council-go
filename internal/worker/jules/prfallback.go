package jules

import (
	"encoding/json"
	"strings"
)

type GHRunner interface {
	Run(argv []string) (stdout, stderr string, exit int, err error)
}

func (l *Live) fallbackPR() {
	if l.Runner == nil || l.Repo == "" || l.sess == nil || l.sess.SID == "" {
		return
	}
	sid := l.sess.SID
	stdout, _, exit, err := l.Runner.Run([]string{"gh", "pr", "list", "--repo", l.Repo, "--state", "open", "--json", "url,headRefName,body"})
	if err != nil || exit != 0 {
		return
	}
	var prs []struct {
		URL         string `json:"url"`
		HeadRefName string `json:"headRefName"`
		Body        string `json:"body"`
	}
	if json.Unmarshal([]byte(stdout), &prs) != nil {
		return
	}
	id := strings.ToLower(sid)
	for _, pr := range prs {
		if pr.URL == "" {
			continue
		}
		body := strings.ToLower(pr.Body)
		if strings.Contains(body, "jules.google.com") && strings.Contains(body, id) {
			l.sess.SetPRURL(pr.URL)
			return
		}
		if strings.HasSuffix(strings.ToLower(pr.HeadRefName), "-"+id) {
			l.sess.SetPRURL(pr.URL)
			return
		}
	}
}
