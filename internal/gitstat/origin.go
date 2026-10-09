package gitstat

import (
	"os/exec"
	"regexp"
	"strings"
)

var originRe = regexp.MustCompile(`github\.com[:/]([^/]+/[^/\s]+?)(?:\.git)?$`)

func OriginRepo(cwd string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	u := strings.TrimSpace(string(out))
	m := originRe.FindStringSubmatch(u)
	if m != nil {
		return m[1]
	}
	return ""
}
