package agy

import (
	"encoding/json"
	"os"
	"os/exec"
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

func TestBuildArgv(t *testing.T) {
	c := cfg.Member{ID: "gemini", CLI: "my-agy", Model: "gemini-pro", Effort: "high"}
	argv := buildArgv(c, "/path/to/workdir")
	want := []string{
		"my-agy", "--sandbox", "--dangerously-skip-permissions",
		"--model", "gemini-pro",
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--add-dir", "/path/to/workdir",
		"--effort", "high",
	}
	if len(argv) != len(want) {
		t.Fatalf("argv len = %d, want %d: %v", len(argv), len(want), argv)
	}
	for i, w := range want {
		if argv[i] != w {
			t.Fatalf("argv[%d] = %q, want %q", i, argv[i], w)
		}
	}
	for _, tok := range argv {
		if tok == "--mode" || tok == "plan" {
			t.Fatalf("argv %v must not contain --mode or plan", argv)
		}
	}
}

func TestHook(t *testing.T) {
	dir := t.TempDir()
	if err := installHook(dir); err != nil {
		t.Fatal(err)
	}
	cfgData, err := os.ReadFile(filepath.Join(dir, ".agents", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var conf map[string]any
	if err := json.Unmarshal(cfgData, &conf); err != nil {
		t.Fatal(err)
	}
	pre, _ := conf["grok-agy-tools"].(map[string]any)["PreToolUse"].([]any)
	hooks, _ := pre[0].(map[string]any)["hooks"].([]any)
	cmdPath, _ := hooks[0].(map[string]any)["command"].(string)
	if fi, err := os.Stat(cmdPath); err != nil || fi.Mode()&0111 == 0 {
		t.Fatalf("hook %q not executable: %v", cmdPath, err)
	}
	for _, tc := range []struct{ n, w string }{{"read_file", "allow"}, {"write_file", "deny"}, {"run_command", "deny"}} {
		cmd := exec.Command(cmdPath)
		cmd.Stdin = strings.NewReader(`{"toolCall":{"name":"` + tc.n + `"}}`)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("exec %s: %v", tc.n, err)
		}
		var resp HookResponse
		if err := json.Unmarshal(out, &resp); err != nil || resp.Decision != tc.w {
			t.Fatalf("%s got %q want %q", tc.n, resp.Decision, tc.w)
		}
	}
	m := New(cfg.Member{ID: "m", CLI: "/nonexistent/bin"}, time.Second)
	if err := m.Start(dir); err == nil || m.hookDir != "" {
		t.Fatalf("Start err=%v hookDir=%q", err, m.hookDir)
	}
}
