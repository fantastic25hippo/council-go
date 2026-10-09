package jules

import (
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type seqDoer struct {
	calls []struct {
		method, path string
		status       int
		body         []byte
	}
	i int
}

func (s *seqDoer) Do(method, path string, body []byte) (int, []byte, error) {
	if s.i >= len(s.calls) {
		return 500, []byte("extra call " + method + " " + path), nil
	}
	c := s.calls[s.i]
	s.i++
	if c.method != "" && c.method != method {
		return 500, []byte("want " + c.method + " got " + method), nil
	}
	if c.path != "" && c.path != path {
		return 500, []byte("want " + c.path + " got " + path), nil
	}
	st := c.status
	if st == 0 {
		st = 200
	}
	return st, c.body, nil
}

func TestLiveStartPollsToCompleted(t *testing.T) {
	api := &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"IN_PROGRESS"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{"activities":[{"planGenerated":{}}]}`)},
		{method: "POST", path: "/sessions/s1:approvePlan", body: []byte(`{}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"COMPLETED","outputs":[{"pullRequest":{"url":"http://pr"}}]}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{"activities":[{"agentMessaged":{"agentMessage":"done"}}]}`)},
	}}
	w := NewLive(&Client{API: api}, time.Minute)
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond
	run := w.Start(worker.Job{Repo: "o/r", Text: "do it"})
	if run.Outcome != worker.OutcomeDone {
		t.Fatalf("outcome %v report %s", run.Outcome, run.Report)
	}
	if run.Raw.PRURL != "http://pr" {
		t.Fatalf("pr %s", run.Raw.PRURL)
	}
	if run.Report != "done" {
		t.Fatalf("report %q", run.Report)
	}
}

func TestLiveStartRequiresRepo(t *testing.T) {
	w := NewLive(&Client{API: &fakeDoer{}}, time.Second)
	run := w.Start(worker.Job{Text: "x"})
	if run.Outcome != worker.OutcomeFailed || !strings.Contains(run.Report, "REPO") {
		t.Fatalf("%v %s", run.Outcome, run.Report)
	}
}

func TestLiveIdleNudge(t *testing.T) {
	api := &seqDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		// 1st poll: IN_PROGRESS, no activities
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"IN_PROGRESS"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
		// Nudge is sent because mock clock jumps by 30 mins
		{method: "POST", path: "/sessions/s1:sendMessage", body: []byte(`{}`)},
		// 2nd poll: COMPLETED
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"COMPLETED"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
	}}
	w := NewLive(&Client{API: api}, time.Hour)
	now := time.Unix(1000, 0)
	w.live.Now = func() time.Time { return now }
	w.live.Sleep = func(d time.Duration) {
		// Jump mock clock by nudge interval to force nudge on first sleep
		now = now.Add(nudgeInterval)
	}
	w.live.Interval = time.Millisecond

	run := w.Start(worker.Job{Repo: "o/r", Text: "do it"})
	if run.Outcome != worker.OutcomeDone {
		t.Fatalf("outcome %v report %s", run.Outcome, run.Report)
	}

	// Verify all calls were consumed, which means the nudge sendMessage was called
	if api.i != len(api.calls) {
		t.Fatalf("not all api calls made, expected %d, got %d", len(api.calls), api.i)
	}
}
