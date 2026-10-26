package agy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var memoCases = []struct {
	name, cmd string
	allow     bool
}{
	{"run_command", "grok-agy-memo wake", true},
	{"run_command", "grok-agy-memo note learned x about y", true},
	{"run_command", "grok-agy-memo wake; rm -rf /", false},
	{"run_command", "rm -rf /", false},
	{"run_command", "cat x | grok-agy-memo wake", false},
	{"run_command", "grok-agy-memo note a $(id)", false},
	{"write_file", "grok-agy-memo wake", false},
	{"view_file", "", true},
}

func TestDecideMemoCommands(t *testing.T) {
	for _, c := range memoCases {
		got, _ := Decide(c.name, c.cmd)
		if (got == "allow") != c.allow {
			t.Errorf("Decide(%q,%q)=%s", c.name, c.cmd, got)
		}
	}
}

func TestPythonHookMatchesGo(t *testing.T) {
	script := filepath.Join(t.TempDir(), "hook.py")
	if err := os.WriteFile(script, []byte(hookScriptContent), 0755); err != nil {
		t.Fatal(err)
	}
	for _, c := range memoCases {
		in := `{"toolCall":{"name":"` + c.name + `","args":{"CommandLine":"` + c.cmd + `"}}}`
		cmd := exec.Command("python3", script)
		cmd.Stdin = strings.NewReader(in)
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), `"allow"`) != c.allow {
			t.Errorf("python hook %q %q -> %s", c.name, c.cmd, out)
		}
	}
}
