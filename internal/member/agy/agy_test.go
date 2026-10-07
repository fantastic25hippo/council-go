package agy

import (
	"os"
	"path/filepath"
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
