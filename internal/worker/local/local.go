package local

import (
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fantastic25hippo/council-go/internal/proc"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// Runner interface allows tests to never exec a real process and never touch the network.
type Runner interface {
	Run(job worker.Job) (stdout string, stderr string, exit int, err error)
}

// Worker is the implementation of worker.Worker.
type Worker struct {
	runner   Runner
	mu       sync.Mutex
	cmd      *exec.Cmd
	canceled bool
}

// New creates a new Worker.
func New(run Runner) *Worker {
	return &Worker{runner: run}
}

// Capabilities returns the worker's capabilities.
func (w *Worker) Capabilities() worker.Capabilities {
	return worker.Capabilities{
		CanAskAndWait:          true,
		KeepsMemoryAcrossFixes: true,
		ReachesGitHub:          false,
		TouchesLocalCheckout:   true,
		OpensPR:                false,
	}
}

// Start initiates a job using the Runner.
func (w *Worker) Start(job worker.Job) worker.Run {
	w.mu.Lock()
	w.canceled = false
	w.mu.Unlock()

	start := time.Now()
	stdout, stderr, exit, err := w.runner.Run(job)
	elapsed := time.Since(start)

	r := worker.Run{
		Raw: worker.Raw{
			ExitCode: exit,
			Stdout:   stdout,
			Stderr:   stderr,
		},
	}

	w.mu.Lock()
	canceled := w.canceled
	w.mu.Unlock()
	if !canceled {
		if c, ok := w.runner.(interface{ IsCanceled() bool }); ok && c.IsCanceled() {
			canceled = true
		}
	}

	if canceled {
		r.Outcome = worker.OutcomeCancelled
		if err != nil {
			r.Report = err.Error()
		} else if stdout != "" {
			r.Report = stdout
		} else {
			r.Report = "cancelled"
		}
		return r
	}

	if err == ErrTimeout {
		r.Outcome = worker.OutcomeTimedOut
		r.Report = timeoutReport(elapsed, checkoutDir(job, w.runner))
		return r
	}

	r.Report = stdout

	if question, asked := worker.ParseQuestion(stdout); asked {
		r.Outcome = worker.OutcomeAsked
		r.Question = question
		return r
	}

	if err != nil {
		r.Outcome = worker.OutcomeFailed
		r.Report = err.Error()
		return r
	}

	if exit == 0 {
		r.Outcome = worker.OutcomeDone
	} else {
		r.Outcome = worker.OutcomeFailed
	}

	return r
}

// Continue builds a continue job from the original job and runs it.
func (w *Worker) Continue(job worker.Job, answer string) worker.Run {
	if ar, ok := w.runner.(*AgyRunner); ok && ar.ConvID != "" {
		job.Text = answer
	} else {
		job.Text = strings.TrimRight(job.Text, "\n") + "\n\n" + answer
	}
	return w.Start(job)
}

// Cancel cancels the local worker.
func (w *Worker) Cancel() error {
	w.mu.Lock()
	w.canceled = true
	cmd := w.cmd
	w.cmd = nil
	runner := w.runner
	w.mu.Unlock()

	var err error
	if cmd != nil {
		err = proc.Kill(cmd)
	}

	if canceler, ok := runner.(interface{ Cancel() error }); ok {
		if cancelErr := canceler.Cancel(); cancelErr != nil {
			if err == nil {
				err = cancelErr
			}
		}
	}

	return err
}
