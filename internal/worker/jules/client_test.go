package jules

import (
	"encoding/json"
	"strings"
	"testing"
)

type fakeDoer struct {
	method string
	path   string
	body   []byte

	status   int
	respBody []byte
	err      error
}

func (f *fakeDoer) Do(method, path string, body []byte) (int, []byte, error) {
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

func TestClient_GetSession(t *testing.T) {
	fake := &fakeDoer{
		respBody: []byte(`{"id":"abc","state":"COMPLETED","outputs":[{"pullRequest":{"url":"http://github.com/pr/1"}}]}`),
	}
	c := &Client{API: fake}

	info, err := c.GetSession("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.method != "GET" {
		t.Errorf("expected GET, got %s", fake.method)
	}
	if fake.path != "/sessions/abc" {
		t.Errorf("expected /sessions/abc, got %s", fake.path)
	}
	if info.ID != "abc" {
		t.Errorf("expected ID abc, got %s", info.ID)
	}
	if info.State != "COMPLETED" {
		t.Errorf("expected State COMPLETED, got %s", info.State)
	}
	if info.PRURL != "http://github.com/pr/1" {
		t.Errorf("expected PR URL http://github.com/pr/1, got %s", info.PRURL)
	}
}

func TestClient_GetSession_EmptyID(t *testing.T) {
	c := &Client{API: &fakeDoer{}}
	_, err := c.GetSession("")
	if err == nil || !strings.Contains(err.Error(), "id") {
		t.Fatalf("expected error containing 'id', got %v", err)
	}
}

func TestClient_GetSession_500(t *testing.T) {
	c := &Client{API: &fakeDoer{status: 500}}
	_, err := c.GetSession("abc")
	if err == nil || !strings.Contains(err.Error(), "jules") {
		t.Fatalf("expected error containing 'jules', got %v", err)
	}
}

func TestClient_GetSession_NameFallback(t *testing.T) {
	fake := &fakeDoer{
		respBody: []byte(`{"name":"sessions/123","state":"RUNNING"}`),
	}
	c := &Client{API: fake}

	info, err := c.GetSession("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ID != "sessions/123" {
		t.Errorf("expected ID sessions/123, got %s", info.ID)
	}
}

func TestClient_Send(t *testing.T) {
	fake := &fakeDoer{}
	c := &Client{API: fake}

	err := c.Send("abc", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.method != "POST" {
		t.Errorf("expected POST, got %s", fake.method)
	}
	if fake.path != "/sessions/abc:sendMessage" {
		t.Errorf("expected /sessions/abc:sendMessage, got %s", fake.path)
	}

	var payload map[string]string
	if err := json.Unmarshal(fake.body, &payload); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if payload["prompt"] != "hello" {
		t.Errorf("expected prompt 'hello', got %s", payload["prompt"])
	}
}

func TestClient_Send_EmptyID(t *testing.T) {
	c := &Client{API: &fakeDoer{}}
	err := c.Send("", "text")
	if err == nil || !strings.Contains(err.Error(), "id") {
		t.Fatalf("expected error containing 'id', got %v", err)
	}
}

func TestClient_Send_500(t *testing.T) {
	c := &Client{API: &fakeDoer{status: 500}}
	err := c.Send("abc", "text")
	if err == nil || !strings.Contains(err.Error(), "jules") {
		t.Fatalf("expected error containing 'jules', got %v", err)
	}
}

func TestClient_ApprovePlan(t *testing.T) {
	fake := &fakeDoer{}
	c := &Client{API: fake}

	err := c.ApprovePlan("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.method != "POST" {
		t.Errorf("expected POST, got %s", fake.method)
	}
	if fake.path != "/sessions/abc:approvePlan" {
		t.Errorf("expected /sessions/abc:approvePlan, got %s", fake.path)
	}
	if string(fake.body) != "{}" {
		t.Errorf("expected body {}, got %s", string(fake.body))
	}
}

func TestClient_ApprovePlan_EmptyID(t *testing.T) {
	c := &Client{API: &fakeDoer{}}
	err := c.ApprovePlan("")
	if err == nil || !strings.Contains(err.Error(), "id") {
		t.Fatalf("expected error containing 'id', got %v", err)
	}
}

func TestClient_ApprovePlan_500(t *testing.T) {
	c := &Client{API: &fakeDoer{status: 500}}
	err := c.ApprovePlan("abc")
	if err == nil || !strings.Contains(err.Error(), "jules") {
		t.Fatalf("expected error containing 'jules', got %v", err)
	}
}
