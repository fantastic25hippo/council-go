package jules

import (
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type planCountDoer struct {
	calls []struct {
		method, path string
		status       int
		body         []byte
	}
	i            int
	approveCount int
}

func (d *planCountDoer) Do(method, path string, body []byte) (int, []byte, error) {
	if strings.Contains(path, ":approvePlan") {
		d.approveCount++
	}
	if d.i >= len(d.calls) {
		return 500, []byte("extra call: " + method + " " + path), nil
	}
	c := d.calls[d.i]
	d.i++
	st := c.status
	if st == 0 {
		st = 200
	}
	return st, c.body, nil
}

func TestLiveAnswerResetsApproved(t *testing.T) {
	api := &planCountDoer{calls: []struct {
		method, path string
		status       int
		body         []byte
	}{
		{method: "POST", path: "/sessions", body: []byte(`{"id":"s1"}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"AWAITING_PLAN_APPROVAL"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
		{method: "POST", path: "/sessions/s1:approvePlan", body: []byte(`{}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"AWAITING_USER_FEEDBACK"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{"activities":[{"agentMessaged":{"agentMessage":"clarify something?"}}]}`)},

		{method: "POST", path: "/sessions/s1:sendMessage", body: []byte(`{}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"AWAITING_PLAN_APPROVAL"}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{}`)},
		{method: "POST", path: "/sessions/s1:approvePlan", body: []byte(`{}`)},
		{method: "GET", path: "/sessions/s1", body: []byte(`{"id":"s1","state":"COMPLETED","outputs":[{"pullRequest":{"url":"http://pr"}}]}`)},
		{method: "GET", path: "/sessions/s1/activities", body: []byte(`{"activities":[{"agentMessaged":{"agentMessage":"all done"}}]}`)},
	}}

	w := NewLive(&Client{API: api}, time.Minute)
	w.live.Sleep = func(time.Duration) {}
	w.live.Interval = time.Millisecond

	job := worker.Job{Repo: "owner/repo", Text: "initial task"}
	run1 := w.Start(job)
	if run1.Outcome != worker.OutcomeAsked {
		t.Fatalf("expected OutcomeAsked, got %v", run1.Outcome)
	}
	if api.approveCount != 1 {
		t.Fatalf("expected 1 ApprovePlan call after first round, got %d", api.approveCount)
	}

	run2 := w.Continue(job, "here is the clarification")
	if run2.Outcome != worker.OutcomeDone {
		t.Fatalf("expected OutcomeDone, got %v", run2.Outcome)
	}
	if api.approveCount != 2 {
		t.Fatalf("expected 2 ApprovePlan calls after answer round, got %d", api.approveCount)
	}
}
