package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func checkoutDir(job worker.Job, runner Runner) string {
	if job.Repo != "" {
		return job.Repo
	}
	if len(job.Paths) > 0 && job.Paths[0] != "" {
		if fi, err := os.Stat(job.Paths[0]); err == nil && !fi.IsDir() {
			return filepath.Dir(job.Paths[0])
		}
		return job.Paths[0]
	}
	if ar, ok := runner.(*AgyRunner); ok {
		return ar.Cwd
	}
	return ""
}

func gitStatusShort(dir string) string {
	if dir == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "status", "--short").Output()
	if err != nil {
		return ""
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return "no files changed"
	}
	lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	if len(lines) > 30 {
		lines = append(lines[:30], fmt.Sprintf("... %d more", len(lines)-30))
	}
	return strings.Join(lines, "\n")
}

func timeoutReport(dur time.Duration, dir string) string {
	s := fmt.Sprintf("worker timed out after %s", dur.Round(time.Millisecond))
	if p := gitStatusShort(dir); p != "" {
		s += "\n" + p
	}
	return s + "\nNext round should do only the smallest remaining slice."
}
