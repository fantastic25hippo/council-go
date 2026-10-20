package agy

import (
	"strings"
	"testing"
)

func TestAgySendDeniedActions(t *testing.T) {
	m, _ := startFake(t, `
while IFS= read -r line; do
  echo '{"event":"step_update","step_update":{"step_type":"agent_response","text_delta":"I tried to run command."}}'
  echo '{"event":"result","result":{"status":"SUCCESS","response":"","denied_actions":[{"display_name":"RunCommand"}]}}'
done
`)
	got, err := m.Send("run something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "[council: agy denied: RunCommand]") {
		t.Fatalf("expected denied message in %q", got)
	}
	if !strings.Contains(got, "I tried to run command.") {
		t.Fatalf("expected response deltas in %q", got)
	}
}

func TestAgySendEmptySuccessReturnsError(t *testing.T) {
	m, _ := startFake(t, `
while IFS= read -r line; do
  echo '{"event":"result","result":{"status":"SUCCESS","response":""}}'
done
`)
	_, err := m.Send("empty")
	if err == nil {
		t.Fatal("expected error for empty SUCCESS, got nil")
	}
	if !strings.Contains(err.Error(), "agy returned an empty response") {
		t.Fatalf("expected 'agy returned an empty response', got %v", err)
	}
}

func TestAgySendMultipleDeniedActions(t *testing.T) {
	m, _ := startFake(t, `
while IFS= read -r line; do
  echo '{"event":"result","result":{"status":"OK","response":"plan here","denied_actions":[{"display_name":"RunCommand"},{"display_name":"WriteFile"}]}}'
done
`)
	got, err := m.Send("plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedDenied := "[council: agy denied: RunCommand, WriteFile]"
	if !strings.Contains(got, expectedDenied) {
		t.Fatalf("expected %q in %q", expectedDenied, got)
	}
	if !strings.Contains(got, "plan here") {
		t.Fatalf("expected plan text in %q", got)
	}
}
