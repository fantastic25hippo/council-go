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
}

func (r *AgyRunner) Run(job worker.Job) (string, string, int, error) {
	f, err := os.CreateTemp("", "council-prompt-*.txt")
	if err != nil {
		return "", "", 1, err
	}
	promptPath := f.Name()
	defer os.Remove(promptPath)

	promptText := job.Text
	r.mu.Lock()
	if r.lastSummary != "" {
		promptText = r.lastSummary + "\n\n" + job.Text
	}
	r.mu.Unlock()

	if _, err := f.WriteString(promptText); err != nil {
		_ = f.Close()
		return "", "", 1, err
	}
	_ = f.Close()

	argv := r.argv(job, promptPath)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = r.Cwd
	cmd.Env = config.StripForbidden(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
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

		jobDir := ""
		for _, line := range strings.Split(stdout.String(), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "job: ") {
				jobDir = strings.TrimSpace(strings.TrimPrefix(line, "job: "))
				break
			}
		}
		if jobDir == "" {
			if err != nil {
				return stdout.String(), stderr.String(), exit, err
			}
			return stdout.String(), stderr.String(), 1, fmt.Errorf("missing job: line in output")
		}

		reportPath := filepath.Join(jobDir, "report.json")
		data, rErr := os.ReadFile(reportPath)
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

func (r *AgyRunner) argv(_ worker.Job, promptFile ...string) []string {
	cli := r.CLI
	if cli == "" {
		cli = "grok-agy-run"
	}
	pf := ""
	if len(promptFile) > 0 {
		pf = promptFile[0]
	}
	effort := r.Effort
	if effort == "" {
		effort = "medium"
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 600 * time.Second
	}
	argv := []string{
		cli, "job",
		"--name", fmt.Sprintf("council-%s", shortID()),
		"--prompt-file", pf,
		"--workdir", r.Cwd,
		"--tools", "code",
		"--effort", effort,
	}
	if r.Model != "" {
		argv = append(argv, "--model", r.Model)
	}
	argv = append(argv, "--wait-secs", fmt.Sprintf("%d", int(timeout.Seconds())))
	return argv
}

func shortID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xffffffff)
	}
	return hex.EncodeToString(b)
}
