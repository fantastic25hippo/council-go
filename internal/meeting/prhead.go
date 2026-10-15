package meeting

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/verify"
)

func parsePRNumber(prURL string) string {
	prURL = strings.TrimSpace(prURL)
	if prURL == "" {
		return ""
	}
	allDigits := true
	for _, r := range prURL {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return prURL
	}
	idx := strings.Index(prURL, "/pull/")
	if idx != -1 {
		rest := prURL[idx+len("/pull/"):]
		var b strings.Builder
		for _, r := range rest {
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
			} else {
				break
			}
		}
		return b.String()
	}
	if strings.HasPrefix(prURL, "pull/") {
		rest := prURL[len("pull/"):]
		var b strings.Builder
		for _, r := range rest {
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
			} else {
				break
			}
		}
		return b.String()
	}
	return ""
}

func gitCmd(dir string, args ...string) *exec.Cmd {
	if dir != "" {
		return exec.Command("git", append([]string{"-C", dir}, args...)...)
	}
	return exec.Command("git", args...)
}

func checkoutPRHead(cwd, prURL string, fn func(dir string) error) error {
	n := parsePRNumber(prURL)
	if n == "" {
		return errors.New("missing PR number")
	}
	ref := fmt.Sprintf("+pull/%s/head:refs/council/pr-%s", n, n)
	councilRef := fmt.Sprintf("refs/council/pr-%s", n)
	if out, err := gitCmd(cwd, "fetch", "origin", ref).CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch: %w: %s", err, strings.TrimSpace(string(out)))
	}
	tmp, err := os.MkdirTemp("", "council-prhead-*")
	if err != nil {
		return err
	}
	defer func() {
		gitCmd(cwd, "worktree", "remove", "--force", tmp).Run()
		os.RemoveAll(tmp)
	}()
	if out, err := gitCmd(cwd, "worktree", "add", "--detach", tmp, councilRef).CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if fn != nil {
		return fn(tmp)
	}
	return nil
}

var verifyPRHead = func(cwd, prURL, cmd string, secs int) (string, int, error) {
	var out string
	var exit int
	err := checkoutPRHead(cwd, prURL, func(dir string) error {
		var runErr error
		out, exit, runErr = verify.Run(dir, cmd, secs)
		return runErr
	})
	return out, exit, err
}
