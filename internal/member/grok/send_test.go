package grok

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
)

const fakeScript = `#!/usr/bin/env python3
import sys, json, time, os

def read_msg():
    length = None
    while True:
        line = sys.stdin.readline()
        if not line:
            return None
        line = line.strip()
        if line.lower().startswith('content-length:'):
            length = int(line.split(':', 1)[1].strip())
        elif line == '' and length is not None:
            return json.loads(sys.stdin.read(length))
        elif line.startswith('{'):
            return json.loads(line)

def send_msg(msg):
    data = json.dumps(msg).encode('utf-8')
    sys.stdout.buffer.write(b"Content-Length: %d\r\n\r\n%s" % (len(data), data))
    sys.stdout.buffer.flush()

mode = "hang" if "hang" in sys.argv else "restart"
count_file = "count.txt"
cnt = 0
if os.path.exists(count_file):
    try:
        with open(count_file, "r") as f:
            cnt = int(f.read().strip())
    except Exception:
        pass
with open(count_file, "w") as f:
    f.write(str(cnt + 1))

while True:
    msg = read_msg()
    if not msg:
        break
    method = msg.get("method")
    mid = msg.get("id")
    if method == "initialize":
        send_msg({"jsonrpc": "2.0", "id": mid, "result": {"protocolVersion": 1}})
    elif method == "session/new":
        send_msg({"jsonrpc": "2.0", "id": mid, "result": {"sessionId": "sid-1"}})
    elif method == "session/prompt":
        if mode == "hang":
            time.sleep(30)
        else:
            if not os.path.exists("state.txt"):
                with open("state.txt", "w") as f:
                    f.write("touched")
                time.sleep(30)
            else:
                with open("state.txt", "a") as f:
                    f.write("-again")
                send_msg({
                    "jsonrpc": "2.0",
                    "method": "session/update",
                    "params": {
                        "update": {
                            "sessionUpdate": "agent_message_chunk",
                            "content": {"text": "hello from fake grok"},
                        }
                    },
                })
                send_msg({"jsonrpc": "2.0", "id": mid, "result": {}})
`

func setupFakeCLI(t *testing.T) {
	t.Helper()
	binDir := t.TempDir()
	fakePath := filepath.Join(binDir, "fake-grok")
	if err := os.WriteFile(fakePath, []byte(fakeScript), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestSendRestartSucceeds(t *testing.T) {
	setupFakeCLI(t)
	cwd := t.TempDir()
	m := New(cfg.Member{ID: "grok", DisplayName: "Grok", CLI: "fake-grok", Model: "restart"}, 200*time.Millisecond)
	m.hangRestartDelay = time.Millisecond
	if err := m.Start(cwd); err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	got, err := m.Send("hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello from fake grok" {
		t.Fatalf("got %q, want %q", got, "hello from fake grok")
	}
}

func TestSendAlwaysHangTimesOut(t *testing.T) {
	setupFakeCLI(t)
	cwd := t.TempDir()
	m := New(cfg.Member{ID: "grok", DisplayName: "Grok", CLI: "fake-grok", Model: "hang"}, 200*time.Millisecond)
	m.hangRestartDelay = time.Millisecond
	if err := m.Start(cwd); err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	_, err := m.Send("hi")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(cwd, "count.txt"))
	if strings.TrimSpace(string(data)) != "3" {
		t.Fatalf("expected 3 attempts, got %s", string(data))
	}
}
