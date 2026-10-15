package jules

import (
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestLiveStartTimedOut(t *testing.T) {
	fake := &fakeDoer{respBody: []byte(`{"id":"s123","state":"IN_PROGRESS"}`)}
	w := NewLive(&Client{API: fake}, 50*time.Millisecond)
	w.live.Interval = time.Millisecond
	w.live.Sleep = func(time.Duration) {}

	run := w.Start(worker.Job{Repo: "o/r", Text: "do it"})
	if run.Outcome != worker.OutcomeTimedOut {
		t.Fatalf("expected OutcomeTimedOut, got %v (report: %s)", run.Outcome, run.Report)
	}
	if !strings.Contains(run.Report, "timed out") {
		t.Fatalf("expected report to contain 'timed out', got %q", run.Report)
	}
	if !strings.Contains(run.Report, "s123") {
		t.Fatalf("expected report to contain session id 's123', got %q", run.Report)
	}
}

func TestLiveNudgeCompletedNoSendMessage(t *testing.T) {
	api := &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"COMPLETED"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
	}}
	w := NewLive(&Client{API: api}, time.Hour)
	w.live.Nudge = time.Millisecond
	w.live.Interval = time.Millisecond
	w.live.Sleep = func(time.Duration) {}
	w.live.Runner = nil

	run := w.Start(worker.Job{Repo: "o/r", Text: "do it"})
	if run.Outcome != worker.OutcomeDone {
		t.Fatalf("expected OutcomeDone, got %v (report: %s)", run.Outcome, run.Report)
	}
	if api.i != len(api.calls) {
		t.Fatalf("expected all %d calls consumed, got %d", len(api.calls), api.i)
	}
}
