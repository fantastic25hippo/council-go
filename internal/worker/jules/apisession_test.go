package jules

import (
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type smallFakeDoer struct {
	method string
	path   string
	body   []byte

	status   int
	respBody []byte
	err      error
}

func (f *smallFakeDoer) Do(method, path string, body []byte) (int, []byte, error) {
	f.method = method
	f.path = path
	f.body = body

	if f.err != nil {
		return f.status, f.respBody, f.err
	}
	if f.status == 0 {
		return 200, f.respBody, nil
	}
	return f.status, f.respBody, nil
}

func TestAPISession_Refresh_EmptyID(t *testing.T) {
	sess := &APISession{
		C:   &Client{API: &smallFakeDoer{}},
		SID: "",
	}
	err := sess.Refresh()
	if err == nil || !strings.Contains(err.Error(), "id") {
		t.Fatalf("expected error containing 'id', got %v", err)
	}
}

func TestAPISession_Refresh_Success(t *testing.T) {
	fake := &smallFakeDoer{
		respBody: []byte(`{"id":"abc","state":"COMPLETED","outputs":[{"pullRequest":{"url":"http://github.com/pr/1"}}]}`),
	}
	c := &Client{API: fake}

	sess := &APISession{
		C:   c,
		SID: "abc",
	}

	err := sess.Refresh()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.method != "GET" {
		t.Errorf("expected GET, got %s", fake.method)
	}
	if fake.path != "/sessions/abc" {
		t.Errorf("expected /sessions/abc, got %s", fake.path)
	}

	if sess.State() != "COMPLETED" {
		t.Errorf("expected State COMPLETED, got %s", sess.State())
	}
	if sess.PRURL() != "http://github.com/pr/1" {
		t.Errorf("expected PR URL http://github.com/pr/1, got %s", sess.PRURL())
	}
	if sess.ID() != "abc" {
		t.Errorf("expected ID abc, got %s", sess.ID())
	}
}

func TestAPISession_Worker_Start(t *testing.T) {
	fake := &smallFakeDoer{
		respBody: []byte(`{"id":"abc","state":"COMPLETED","outputs":[{"pullRequest":{"url":"http://github.com/pr/1"}}]}`),
	}
	c := &Client{API: fake}

	sess := &APISession{
		C:   c,
		SID: "abc",
	}

	err := sess.Refresh()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w := New(sess)
	run := w.Start(worker.Job{})

	if run.Outcome != worker.OutcomeDone {
		t.Errorf("expected OutcomeDone, got %v", run.Outcome)
	}
	if run.Raw.SessionID != "abc" {
		t.Errorf("expected SessionID 'abc', got %q", run.Raw.SessionID)
	}
	if run.Raw.PRURL != "http://github.com/pr/1" {
		t.Errorf("expected PRURL 'http://github.com/pr/1', got %q", run.Raw.PRURL)
	}
}
