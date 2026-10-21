package grok

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
)

const fakeInitScript = `#!/usr/bin/env python3
import sys, json

def read_msg():
    line = sys.stdin.readline()
    if not line:
        return None
    return json.loads(line.strip())

def send_msg(msg):
    data = json.dumps(msg) + "\n"
    sys.stdout.write(data)
    sys.stdout.flush()

while True:
    msg = read_msg()
    if not msg:
        break
    method = msg.get("method")
    mid = msg.get("id")
    if method == "initialize":
        with open("init_params.json", "w") as f:
            json.dump(msg.get("params", {}), f)
        send_msg({"jsonrpc": "2.0", "id": mid, "result": {"protocolVersion": 1}})
    elif method == "session/new":
        send_msg({"jsonrpc": "2.0", "id": mid, "result": {"sessionId": "sid-init"}})
`

func TestInitializeCapturedParams(t *testing.T) {
	cwd := t.TempDir()
	fakePath := filepath.Join(cwd, "fake-grok-init")
	if err := os.WriteFile(fakePath, []byte(fakeInitScript), 0755); err != nil {
		t.Fatal(err)
	}

	m := New(cfg.Member{ID: "grok", DisplayName: "Grok", CLI: fakePath, Model: "test"}, 2*time.Second)
	if err := m.Start(cwd); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer m.Close()

	paramBytes, err := os.ReadFile(filepath.Join(cwd, "init_params.json"))
	if err != nil {
		t.Fatalf("failed to read captured params: %v", err)
	}
	var params map[string]any
	if err := json.Unmarshal(paramBytes, &params); err != nil {
		t.Fatalf("failed to unmarshal params: %v", err)
	}

	caps, ok := params["clientCapabilities"].(map[string]any)
	if !ok {
		t.Fatalf("missing or invalid clientCapabilities: %v", params)
	}

	fsCaps, ok := caps["fs"].(map[string]any)
	if !ok {
		t.Fatalf("missing or invalid fs capabilities: %v", caps)
	}
	if fsCaps["readTextFile"] != true {
		t.Fatalf("expected fs.readTextFile == true, got %v", fsCaps["readTextFile"])
	}
	if fsCaps["writeTextFile"] != false {
		t.Fatalf("expected fs.writeTextFile == false, got %v", fsCaps["writeTextFile"])
	}
	if caps["terminal"] != false {
		t.Fatalf("expected terminal == false, got %v", caps["terminal"])
	}
}

func TestStartClosesExistingSubprocess(t *testing.T) {
	cwd := t.TempDir()
	fakePath := filepath.Join(cwd, "fake-grok-init")
	if err := os.WriteFile(fakePath, []byte(fakeInitScript), 0755); err != nil {
		t.Fatal(err)
	}

	m := New(cfg.Member{ID: "grok", DisplayName: "Grok", CLI: fakePath, Model: "test"}, 2*time.Second)
	if err := m.Start(cwd); err != nil {
		t.Fatalf("First Start failed: %v", err)
	}

	cmd1 := m.cmd
	if cmd1 == nil || cmd1.Process == nil {
		t.Fatal("expected non-nil cmd1")
	}

	if err := m.Start(cwd); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}
	defer m.Close()

	cmd2 := m.cmd
	if cmd2 == nil || cmd2.Process == nil {
		t.Fatal("expected non-nil cmd2")
	}

	if cmd1 == cmd2 {
		t.Fatal("expected new cmd process on second Start")
	}

	// Verify cmd1 process is no longer running
	if err := cmd1.Process.Signal(syscall.Signal(0)); err == nil {
		t.Fatal("expected error signaling terminated cmd1 process, got nil")
	}
}
