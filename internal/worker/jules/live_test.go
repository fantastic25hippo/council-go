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
