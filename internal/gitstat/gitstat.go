package gitstat

import (
	"os/exec"
	"strings"
)

var Fetcher = func(dir string) error {
	_, err := git(dir, "fetch", "--quiet")
	return err
}

// State reports dirty tracked files and unpushed commits. Fail-closed: error
// if git cannot answer.
func State(dir string) (dirty, unpushed bool, err error) {
	if err := Fetcher(dir); err != nil {
		return false, false, err
	}
	status, err := git(dir, "status", "--porcelain")
	if err != nil {
		return false, false, err
	}
	dirty = strings.TrimSpace(status) != ""
	// unpushed: commits not on @{u}. If no upstream, treat as unpushed.
	_, err = git(dir, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		return dirty, true, nil
	}
	ahead, err := git(dir, "rev-list", "--count", "@{u}..HEAD")
	if err != nil {
		return dirty, false, err
	}
	unpushed = strings.TrimSpace(ahead) != "0"
	return dirty, unpushed, nil
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
