package meeting_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/meeting"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

type cancelBlockingWorker struct {
	caps    worker.Capabilities
	entered chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	cancels int
}

func newCancelBlockingWorker(caps worker.Capabilities) *cancelBlockingWorker {
	return &cancelBlockingWorker{
		caps:    caps,
		entered: make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (w *cancelBlockingWorker) Capabilities() worker.Capabilities {
	return w.caps
}

func (w *cancelBlockingWorker) Start(job worker.Job) worker.Run {
	select {
	case <-w.entered:
	default:
		close(w.entered)
	}
	<-w.done
	return worker.Run{Outcome: worker.OutcomeCancelled, Report: "cancelled"}
}

func (w *cancelBlockingWorker) Continue(job worker.Job, answer string) worker.Run {
	return worker.Run{Outcome: worker.OutcomeCancelled, Report: "cancelled"}
}

func (w *cancelBlockingWorker) Cancel() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cancels++
	select {
	case <-w.done:
	default:
		close(w.done)
	}
	return nil
}

func TestRunCancelInFlightWorker(t *testing.T) {
	a := &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{
		"Hello, I am Grok.",
		agreedPlan(),
		"VERDICT: PASS\n- looks good",
	}}
	b := &member.Fake{Name: "gemini", Shown: "Gemini", Replies: []string{
		"Hello, I am Gemini.",
		"local is fine",
		"LAUNCH: local\nAGREED",
		"VERDICT: PASS\n- ok",
	}}
	blockW := newCancelBlockingWorker(worker.Capabilities{
		TouchesLocalCheckout: true,
		CanAskAndWait:        true,
	})
	var out bytes.Buffer
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:     cfg.Config{CheckUnpushed: false, DebateTurnsEach: 3, FixRoundsMax: 3, Worker: "local"},
		A:       a,
		B:       b,
		IO:      &term.IO{In: strings.NewReader("hi\nfix the foo\n\nquit\n"), Out: &out},
		Council: council,
		Relay:   &relay.Relay{Council: council, Worker: blockW},
		WorkerFor: func(id string) worker.Worker {
			if id != "local" {
				t.Errorf("worker id %s", id)
			}
			return blockW
		},
		Cwd: t.TempDir(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Run(ctx)
	}()

	select {
	case <-blockW.entered:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker Start to be entered")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Session.Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Session.Run did not return within timeout after cancellation")
	}

	blockW.mu.Lock()
	cancels := blockW.cancels
	blockW.mu.Unlock()

	if cancels == 0 {
		t.Fatalf("expected Cancel to have been called on worker, got %d", cancels)
	}
}
