package jules

import (
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestLivePausedEndsPoll(t *testing.T) {
	api := &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"PAUSED"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
	}}
	w := NewLive(&Client{API: api}, time.Hour)
	w.live.Runner = nil
	w.live.Branch = "master"
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond
	var capturedWarn string
	w.live.warn = func(msg string) {
		capturedWarn = msg
	}
	run := w.Start(worker.Job{Repo: "o/r", Text: "do it", Branch: "master"})
	if run.Outcome != worker.OutcomeAsked {
		t.Fatalf("expected OutcomeAsked, got %v", run.Outcome)
	}
	if run.Question != "Jules session is paused" {
		t.Fatalf("expected question %q, got %q", "Jules session is paused", run.Question)
	}
	if !strings.Contains(capturedWarn, "PAUSED") {
		t.Fatalf("expected captured warn to contain PAUSED, got %q", capturedWarn)
	}
}

func TestLiveCancelledEndsPoll(t *testing.T) {
	api := &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"CANCELLED"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
	}}
	w := NewLive(&Client{API: api}, time.Hour)
	w.live.Runner = nil
	w.live.Branch = "master"
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond
	run := w.Start(worker.Job{Repo: "o/r", Text: "do it", Branch: "master"})
	if run.Outcome != worker.OutcomeCancelled {
		t.Fatalf("expected OutcomeCancelled, got %v", run.Outcome)
	}
}

func TestWorkerMapPausedCancelled(t *testing.T) {
	w1 := New(&fakeSession{state: "PAUSED", msg: "x"})
	run1 := w1.Start(worker.Job{})
	if run1.Outcome != worker.OutcomeAsked {
		t.Fatalf("expected OutcomeAsked, got %v", run1.Outcome)
	}
	if run1.Question != "Jules session is paused" {
		t.Fatalf("expected question %q, got %q", "Jules session is paused", run1.Question)
	}

	w2 := New(&fakeSession{state: "CANCELLED"})
	run2 := w2.Start(worker.Job{})
	if run2.Outcome != worker.OutcomeCancelled {
		t.Fatalf("expected OutcomeCancelled, got %v", run2.Outcome)
	}
}
