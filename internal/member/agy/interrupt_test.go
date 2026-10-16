package agy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
)

const liveInterrupt = "The stream was interrupted. Please continue the task you were working on."

func startFake(t *testing.T, body string) (*Member, string) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ncd \""+dir+"\"\n"+body+"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	m := New(cfg.Member{ID: "gemini", DisplayName: "Gemini", CLI: script, Model: "m"}, 5*time.Second)
	if err := m.Start(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m, dir
}

func stdinLog(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "in.log"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func countWrites(log string) int {
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func TestAgyInterruptWithDeltasReturnsText(t *testing.T) {
	m, dir := startFake(t, `
while IFS= read -r line; do
  echo "$line" >> in.log
  echo '{"event":"step_update","step_update":{"step_type":"agent_response","text_delta":"hello gemini"}}'
  echo '{"event":"result","result":{"status":"ERROR","error":"`+liveInterrupt+`","response":""}}'
done
`)
	got, err := m.Send("greet")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello gemini" {
		t.Fatalf("got %q", got)
	}
	log := stdinLog(t, dir)
	if countWrites(log) != 1 {
		t.Fatalf("writes %d log %q", countWrites(log), log)
	}
	if strings.Contains(log, streamContinuePrompt) {
		t.Fatalf("sent continue: %q", log)
	}
}

func TestAgyInterruptEmptyThenContinue(t *testing.T) {
	m, dir := startFake(t, `
n=0
while IFS= read -r line; do
  echo "$line" >> in.log
  n=$((n+1))
  if [ "$n" -eq 1 ]; then
    echo '{"event":"result","result":{"status":"ERROR","error":"`+liveInterrupt+`","response":""}}'
  else
    echo '{"event":"result","result":{"response":"continued hi","status":"OK"}}'
  fi
done
`)
	got, err := m.Send("greet")
	if err != nil {
		t.Fatal(err)
	}
	if got != "continued hi" {
		t.Fatalf("got %q", got)
	}
	log := stdinLog(t, dir)
	if countWrites(log) != 2 {
		t.Fatalf("writes %d log %q", countWrites(log), log)
	}
	if !strings.Contains(log, streamContinuePrompt) {
		t.Fatalf("missing continue: %q", log)
	}
}

func TestAgyInterruptContinueCap(t *testing.T) {
	m, dir := startFake(t, `
while IFS= read -r line; do
  echo "$line" >> in.log
  echo '{"event":"result","result":{"status":"ERROR","error":"`+liveInterrupt+`","response":""}}'
done
`)
	_, err := m.Send("greet")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "stream was interrupted") {
		t.Fatalf("err %v", err)
	}
	log := stdinLog(t, dir)
	if countWrites(log) != 1+maxStreamContinues {
		t.Fatalf("writes %d log %q", countWrites(log), log)
	}
}

func TestAgyNonInterruptError(t *testing.T) {
	m, dir := startFake(t, `
while IFS= read -r line; do
  echo "$line" >> in.log
  echo '{"event":"result","result":{"status":"ERROR","error":"quota exceeded","response":""}}'
done
`)
	_, err := m.Send("greet")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("err %v", err)
	}
	log := stdinLog(t, dir)
	if countWrites(log) != 1 {
		t.Fatalf("writes %d log %q", countWrites(log), log)
	}
}
