package jules

import (
	"context"
	"errors"

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
	live *Live
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
	if w.live != nil {
		if err := w.live.start(job); err != nil {
			if errors.Is(err, context.Canceled) {
				return worker.Run{Outcome: worker.OutcomeCancelled, Report: "cancelled"}
			}
			return worker.Run{Outcome: worker.OutcomeFailed, Report: err.Error()}
		}
		w.sess = w.live.session()
	}
	return w.mapStateToRun()
}

func (w *Worker) Continue(job worker.Job, answer string) worker.Run {
	if recorder, ok := w.sess.(interface{ Record(string) }); ok {
		recorder.Record(answer)
	}
	if w.live != nil {
		if err := w.live.answer(answer); err != nil {
			if errors.Is(err, context.Canceled) {
				return worker.Run{Outcome: worker.OutcomeCancelled, Report: "cancelled"}
			}
			return worker.Run{Outcome: worker.OutcomeFailed, Report: err.Error()}
		}
		w.sess = w.live.session()
	}
	return w.mapStateToRun()
}

func (w *Worker) mapStateToRun() worker.Run {
	if w.sess == nil {
		return worker.Run{Outcome: worker.OutcomeFailed, Report: "no session"}
	}
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
	if w.live != nil && w.live.cancel != nil {
		w.live.cancel()
	}
	return nil
}
