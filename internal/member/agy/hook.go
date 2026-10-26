package agy

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type HookResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

var allowedTools = map[string]bool{
	"view_file":    true,
	"read_file":    true,
	"list_dir":     true,
	"grep_search":  true,
	"find_by_name": true,
}

// memoCmd matches only "grok-agy-memo wake" or "grok-agy-memo note ..." optionally after "cd DIR && ", with no other shell syntax.
var memoCmd = regexp.MustCompile("^(cd [A-Za-z0-9_./-]+ && )?grok-agy-memo (wake|note [^;&|`$<>()\\n]*)$")

func Decide(name, command string) (string, string) {
	if allowedTools[name] || (name == "run_command" && memoCmd.MatchString(strings.TrimSpace(command))) {
		return "allow", ""
	}
	return "deny", fmt.Sprintf("tool %s is not permitted for read-only member", name)
}

func extractToolName(payload map[string]any) string {
	if tc, ok := payload["toolCall"].(map[string]any); ok {
		if n, ok := tc["name"].(string); ok {
			return n
		}
	}
	if tc, ok := payload["tool_call"].(map[string]any); ok {
		if n, ok := tc["name"].(string); ok {
			return n
		}
	}
	if n, ok := payload["name"].(string); ok {
		return n
	}
	return ""
}

// findCommand returns the first string under a command-like key, at any depth.
func findCommand(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range []string{"CommandLine", "commandLine", "command", "Command"} {
			if c, ok := t[k].(string); ok {
				return c
			}
		}
		for _, x := range t {
			if c := findCommand(x); c != "" {
				return c
			}
		}
	}
	return ""
}

func HandleHook(r io.Reader, w io.Writer) error {
	var payload map[string]any
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return json.NewEncoder(w).Encode(HookResponse{
			Decision: "deny",
			Reason:   "invalid json: " + err.Error(),
		})
	}
	name := extractToolName(payload)
	decision, reason := Decide(name, findCommand(payload))
	return json.NewEncoder(w).Encode(HookResponse{
		Decision: decision,
		Reason:   reason,
	})
}

const hookScriptContent = `#!/usr/bin/env python3
import sys, json, re

try:
    data = json.load(sys.stdin)
except Exception:
    data = {}

tc = data.get("toolCall") or data.get("tool_call") or data
name = str(tc.get("name") if isinstance(tc, dict) else "")

allowed = {"view_file", "read_file", "list_dir", "grep_search", "find_by_name"}

def find_command(v):
    if isinstance(v, dict):
        for k in ("CommandLine", "commandLine", "command", "Command"):
            if isinstance(v.get(k), str):
                return v[k]
        for x in v.values():
            c = find_command(x)
            if c:
                return c
    return ""

memo = re.match(r"^(cd [A-Za-z0-9_./-]+ && )?grok-agy-memo (wake|note [^;&|$<>()\n\x60]*)$", find_command(data).strip())

if name in allowed or (name == "run_command" and memo):
    res = {"decision": "allow"}
else:
    res = {"decision": "deny", "reason": f"tool {name} is not permitted for read-only member"}

print(json.dumps(res))
`

func installHook(dir string) error {
	agentsDir := filepath.Join(dir, ".agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return err
	}
	hookScript := filepath.Join(dir, "hook.py")
	absScript, err := filepath.Abs(hookScript)
	if err == nil {
		hookScript = absScript
	}
	if err := os.WriteFile(hookScript, []byte(hookScriptContent), 0755); err != nil {
		return err
	}
	hooksConfig := fmt.Sprintf(`{
  "grok-agy-tools": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "%s",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
`, hookScript)
	return os.WriteFile(filepath.Join(agentsDir, "hooks.json"), []byte(hooksConfig), 0644)
}
