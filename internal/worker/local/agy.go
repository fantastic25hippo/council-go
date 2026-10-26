package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fantastic25hippo/council-go/internal/config"
	"github.com/fantastic25hippo/council-go/internal/proc"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// ErrTimeout is returned when the local process exceeds Timeout.
var ErrTimeout = fmt.Errorf("worker timed out")

// AgyRunner runs one-shot `agy -p` jobs. -p is last.
type AgyRunner struct {
	CLI      string
	Model    string
	Effort   string
	Cwd      string
	Timeout  time.Duration
	ConvID   string
	mu       sync.Mutex
	cmd      *exec.Cmd
	canceled bool
}

func (r *AgyRunner) Run(job worker.Job) (string, string, int, error) {
	argv := r.argv(job)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = r.Cwd
	cmd.Env = config.ChildEnv(os.Environ(), true)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", "", 1, err
	}
	r.mu.Lock()
	r.cmd = cmd
	r.canceled = false
	r.mu.Unlock()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 600 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-done:
		r.mu.Lock()
		r.cmd = nil
		r.mu.Unlock()
		exit := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exit = ee.ExitCode()
			} else {
				return stdout.String(), stderr.String(), 1, err
			}
		}

		outText := stdout.String()
		var js map[string]interface{}
		if json.Unmarshal([]byte(outText), &js) == nil {
			if cid, ok := js["conversation_id"].(string); ok && cid != "" {
				r.ConvID = cid
			} else if res, ok := js["result"].(map[string]interface{}); ok {
				if cid, ok := res["conversation_id"].(string); ok && cid != "" {
					r.ConvID = cid
				}
			}
			outText = parseReport(js, outText, stderr.String())
		}

		return outText, stderr.String(), exit, nil
	case <-timer.C:
		_ = proc.Kill(cmd)
		<-done
		r.mu.Lock()
		r.cmd = nil
		r.mu.Unlock()
		return stdout.String(), stderr.String(), 1, ErrTimeout
	}
}

func (r *AgyRunner) Cancel() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canceled = true
	if r.cmd == nil {
		return nil
	}
	err := proc.Kill(r.cmd)
	r.cmd = nil
	return err
}

func (r *AgyRunner) IsCanceled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.canceled
}

func (r *AgyRunner) argv(job worker.Job) []string {
	cli := r.CLI
	if cli == "" {
		cli = "agy"
	}
	argv := []string{cli}
	if r.Model != "" {
		argv = append(argv, "--model", r.Model)
	}
	if r.Effort != "" {
		argv = append(argv, "--effort", r.Effort)
	}
	for _, d := range addDirs(r.Cwd, job.Paths) {
		argv = append(argv, "--add-dir", d)
	}
	if r.ConvID != "" {
		argv = append(argv, "--conversation", r.ConvID)
	}
	argv = append(argv, "--dangerously-skip-permissions")
	argv = append(argv, "--output-format", "json")
	argv = append(argv, "-p", job.Text)
	return argv
}

func addDirs(cwd string, paths []string) []string {
	var dirs []string
	seen := map[string]bool{}
	add := func(d string) {
		if seen[d] {
			return
		}
		seen[d] = true
		dirs = append(dirs, d)
	}
	for _, p := range paths {
		abs := p
		if !filepath.IsAbs(p) {
			abs = filepath.Join(cwd, p)
		}
		fi, err := os.Stat(abs)
		if err != nil {
			continue
		}
		if !fi.IsDir() {
			abs = filepath.Dir(abs)
		}
		add(abs)
	}
	if len(dirs) == 0 && cwd != "" {
		add(cwd)
	}
	return dirs
}
