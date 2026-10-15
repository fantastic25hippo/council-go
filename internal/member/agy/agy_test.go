package agy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
)

func TestAgySendCollectsResult(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agy")
	body := "#!/bin/sh\n" +
		"while IFS= read -r line; do\n" +
		"  echo '{\"event\":\"step_update\",\"step_update\":{\"step_type\":\"agent_response\",\"text_delta\":\"hi \"}}'\n" +
		"  echo '{\"event\":\"result\",\"result\":{\"response\":\"hello from agy\",\"status\":\"OK\"}}'\n" +
		"done\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	m := New(cfg.Member{ID: "gemini", DisplayName: "Gemini", CLI: script, Model: "m"}, 5*time.Second)
	if err := m.Start(dir); err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	got, err := m.Send("hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello from agy" {
		t.Fatalf("got %q", got)
	}
}

func TestAgySendHangThenRestart(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state")
	script := filepath.Join(dir, "fake-agy")
	body := "#!/bin/sh\n" +
		"if [ ! -f \"" + stateFile + "\" ]; then\n" +
		"  touch \"" + stateFile + "\"\n" +
		"  sleep 60\n" +
		"else\n" +
		"  while IFS= read -r line; do\n" +
		"    echo '{\"event\":\"step_update\",\"step_update\":{\"step_type\":\"agent_response\",\"text_delta\":\"hi \"}}'\n" +
		"    echo '{\"event\":\"result\",\"result\":{\"response\":\"hello from agy\",\"status\":\"OK\"}}'\n" +
		"  done\n" +
		"fi\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	m := New(cfg.Member{ID: "gemini", DisplayName: "Gemini", CLI: script, Model: "m"}, 200*time.Millisecond)
	m.hangRestartDelay = 1 * time.Millisecond
	if err := m.Start(dir); err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	got, err := m.Send("hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello from agy" {
		t.Fatalf("got %q", got)
	}
}

func TestAgySendAlwaysHang(t *testing.T) {
	dir := t.TempDir()
	runsFile := filepath.Join(dir, "runs")
	script := filepath.Join(dir, "fake-agy")
	body := "#!/bin/sh\n" +
		"echo 1 >> \"" + runsFile + "\"\n" +
		"sleep 60\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	m := New(cfg.Member{ID: "gemini", DisplayName: "Gemini", CLI: script, Model: "m"}, 50*time.Millisecond)
	m.hangRestartDelay = 1 * time.Millisecond
	if err := m.Start(dir); err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	_, err := m.Send("hello")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected error containing timeout, got %v", err)
	}
	data, err := os.ReadFile(runsFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 runs (1 initial + 2 restarts), got %d", len(lines))
	}
}
