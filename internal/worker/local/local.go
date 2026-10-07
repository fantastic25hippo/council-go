package local

import (
	"strings"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

// Runner interface allows tests to never exec a real process and never touch the network.
type Runner interface {
	Run(job worker.Job) (stdout string, stderr string, exit int, err error)
}

// Worker is the implementation of worker.Worker.
type Worker struct {
	runner Runner
}

// New creates a new Worker.
func New(run Runner) *Worker {
	return &Worker{runner: run}
}

// Capabilities returns the worker's capabilities.
func (w *Worker) Capabilities() worker.Capabilities {
	return worker.Capabilities{
		CanAskAndWait:          true,
		KeepsMemoryAcrossFixes: false,
		ReachesGitHub:          false,
		TouchesLocalCheckout:   true,
		OpensPR:                false,
	}
}

// Start initiates a job using the Runner.
func (w *Worker) Start(job worker.Job) worker.Run {
	stdout, stderr, exit, err := w.runner.Run(job)

	r := worker.Run{
		Raw: worker.Raw{
			ExitCode: exit,
			Stdout:   stdout,
			Stderr:   stderr,
		},
	}

	if err != nil {
		r.Outcome = worker.OutcomeFailed
		r.Report = err.Error()
		return r
	}

	r.Report = stdout

	question, asked := worker.ParseQuestion(stdout)
	if asked {
		r.Outcome = worker.OutcomeAsked
		r.Question = question
	} else if exit == 0 {
		r.Outcome = worker.OutcomeDone
	} else {
		r.Outcome = worker.OutcomeFailed
	}

	return r
}

// Continue builds a continue job from the original job and runs it.
func (w *Worker) Continue(job worker.Job, answer string) worker.Run {
	var keep []string
	lines := strings.Split(job.Text, "\n")
	for _, line := range lines {
		if strings.Contains(line, "VERIFICATION FAILURE") || strings.Contains(line, "fix reasons") {
			break
		}
		keep = append(keep, line)
	}

	base := strings.Join(keep, "\n")
	job.Text = base + "\n\nANSWER:" + answer
	return w.Start(job)
}

// Cancel cancels the local worker.
func (w *Worker) Cancel() error {
	return nil
}
