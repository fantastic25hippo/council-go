package jules

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreateSession(t *testing.T) {
	fake := &fakeDoer{respBody: []byte(`{"id":"abc","state":"IN_PROGRESS"}`)}
	c := &Client{API: fake}
	id, err := c.CreateSession("owner/repo", "do the thing", "master", "Council task")
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc" {
		t.Fatalf("id %s", id)
	}
	if fake.method != "POST" || fake.path != "/sessions" {
		t.Fatalf("%s %s", fake.method, fake.path)
	}
	var payload map[string]any
	if err := json.Unmarshal(fake.body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["automationMode"] != "AUTO_CREATE_PR" {
		t.Fatalf("%v", payload)
	}
	if payload["requirePlanApproval"] != false {
		t.Fatalf("plan approval %v", payload["requirePlanApproval"])
	}
}

func TestCreateSessionNamePrefix(t *testing.T) {
	fake := &fakeDoer{respBody: []byte(`{"name":"sessions/xyz"}`)}
	c := &Client{API: fake}
	id, err := c.CreateSession("o/r", "p", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if id != "xyz" {
		t.Fatalf("id %s", id)
	}
}

func TestCreateSessionEmptyRepo(t *testing.T) {
	c := &Client{API: &fakeDoer{}}
	if _, err := c.CreateSession("", "p", "", ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateSessionHTTPError(t *testing.T) {
	c := &Client{API: &fakeDoer{status: 500, respBody: []byte("no")}}
	_, err := c.CreateSession("o/r", "p", "", "")
	if err == nil || !strings.Contains(err.Error(), "jules") {
		t.Fatalf("%v", err)
	}
}
