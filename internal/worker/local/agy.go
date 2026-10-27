package local

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fantastic25hippo/council-go/internal/config"
	"github.com/fantastic25hippo/council-go/internal/proc"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// ErrTimeout is returned when the local process exceeds Timeout.
var ErrTimeout = fmt.Errorf("worker timed out")

// AgyRunner runs one-shot grok-agy-run job commands.
type AgyRunner struct {
	CLI         string
	Model       string
	Effort      string
	Cwd         string
	Timeout     time.Duration
	ConvID      string
	lastSummary string
	mu          sync.Mutex
	cmd         *exec.Cmd
	canceled    bool
	afterStart  func()
}

func (r *AgyRunner) Run(job worker.Job) (string, string, int, error) {
	f, err := os.CreateTemp("", "council-prompt-*.txt")
	if err != nil {
		return "", "", 1, err
	}
	defer os.Remove(f.Name())

	promptText := job.Text
	r.mu.Lock()
	if r.lastSummary != "" {
		promptText = r.lastSummary + "\n\n" + job.Text
	}
	r.canceled = false
	r.mu.Unlock()

	_, wErr := f.WriteString(promptText)
	_ = f.Close()
	if wErr != nil {
		return "", "", 1, wErr
	}

	argv := r.argv(job, f.Name())
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir, cmd.Env = r.Cwd, config.StripForbidden(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		return "", "", 1, err
	}
	if r.afterStart != nil {
		r.afterStart()
	}
	if !r.attachCmd(cmd) {
		_ = cmd.Wait()
		return stdout.String(), stderr.String(), 1, fmt.Errorf("canceled")
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(r.timeout())
	defer timer.Stop()
	select {
	case err := <-done:
		r.mu.Lock()
		r.cmd = nil
		r.mu.Unlock()
		exit := 0
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else if err != nil {
			return stdout.String(), stderr.String(), 1, err
		}

		jobDir := ""
		for _, line := range strings.Split(stdout.String(), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "job: ") {
				jobDir = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "job: "))
				break
			}
		}
		if jobDir == "" {
			if err != nil {
				return stdout.String(), stderr.String(), exit, err
			}
			return stdout.String(), stderr.String(), 1, fmt.Errorf("missing job: line in output")
		}

		data, rErr := os.ReadFile(filepath.Join(jobDir, "report.json"))
		if rErr != nil {
			return stdout.String(), stderr.String(), 1, fmt.Errorf("reading report.json: %w", rErr)
		}
		var rep struct {
			Status       string `json:"status"`
			FinalSummary string `json:"final_summary"`
		}
		if err := json.Unmarshal(data, &rep); err != nil {
			return stdout.String(), stderr.String(), 1, fmt.Errorf("parsing report.json: %w", err)
		}
		if rep.Status != "ok" {
			if exit == 0 {
				exit = 1
			}
			return rep.FinalSummary, stderr.String(), exit, fmt.Errorf("job failed with status %q", rep.Status)
		}
		r.mu.Lock()
		r.lastSummary = rep.FinalSummary
		r.mu.Unlock()
		return rep.FinalSummary, stderr.String(), exit, nil
	case <-timer.C:
		_ = proc.Kill(cmd)
		<-done
		r.mu.Lock()
		r.cmd = nil
		r.mu.Unlock()
		return stdout.String(), stderr.String(), 1, ErrTimeout
	}
}

func (r *AgyRunner) Cancel() (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canceled = true
	if r.cmd != nil {
		err = proc.Kill(r.cmd)
		r.cmd = nil
	}
	return err
}

func (r *AgyRunner) IsCanceled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.canceled
}

func (r *AgyRunner) attachCmd(cmd *exec.Cmd) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.canceled {
		_ = proc.Kill(cmd)
		return false
	}
	r.cmd = cmd
	return true
}

func (r *AgyRunner) timeout() time.Duration {
	if r.Timeout > 0 {
		return r.Timeout
	}
	return 600 * time.Second
}

func (r *AgyRunner) argv(_ worker.Job, promptFile ...string) []string {
	cli, pf, effort := r.CLI, "", r.Effort
	if cli == "" {
		cli = "grok-agy-run"
	}
	if len(promptFile) > 0 {
		pf = promptFile[0]
	}
	if effort == "" {
		effort = "medium"
	}
	argv := []string{
		cli, "job", "--name", fmt.Sprintf("council-%s", shortID()),
		"--prompt-file", pf, "--workdir", r.Cwd, "--tools", "code", "--effort", effort,
	}
	if r.Model != "" {
		argv = append(argv, "--model", r.Model)
	}
	return append(argv, "--wait-secs", fmt.Sprintf("%d", int(r.timeout().Seconds())))
}

func shortID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
