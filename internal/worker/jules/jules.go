package jules

import (
	"github.com/fantastic25hippo/council-go/internal/worker"
)

type Session interface {
	State() string
	LastAgentMessage() string
	ID() string
	PRURL() string
}

type Worker struct {
	sess Session
}

func New(sess Session) *Worker {
	return &Worker{sess: sess}
}

func (w *Worker) Capabilities() worker.Capabilities {
	return worker.Capabilities{
		CanAskAndWait:          true,
		KeepsMemoryAcrossFixes: true,
		ReachesGitHub:          true,
		TouchesLocalCheckout:   false,
		OpensPR:                true,
	}
}

func (w *Worker) Start(job worker.Job) worker.Run {
	return w.mapStateToRun()
}

func (w *Worker) Continue(job worker.Job, answer string) worker.Run {
	if recorder, ok := w.sess.(interface{ Record(string) }); ok {
		recorder.Record(answer)
	}
	return w.mapStateToRun()
}

func (w *Worker) mapStateToRun() worker.Run {
	state := w.sess.State()
	msg := w.sess.LastAgentMessage()

	run := worker.Run{
		Raw: worker.Raw{
			SessionID: w.sess.ID(),
			PRURL:     w.sess.PRURL(),
		},
	}

	switch state {
	case "COMPLETED":
		run.Outcome = worker.OutcomeDone
		run.Report = msg
	case "AWAITING_USER_FEEDBACK":
		run.Outcome = worker.OutcomeAsked
		run.Question = msg
		run.Report = msg
	case "FAILED":
		run.Outcome = worker.OutcomeFailed
		run.Report = msg
	default:
		run.Outcome = worker.OutcomeFailed
		run.Report = "unusable state: " + state
	}

	return run
}

func (w *Worker) Cancel() error {
	return nil
}
