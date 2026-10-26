package agy

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func Decide(name string) (string, string) {
	if allowedTools[name] {
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

func HandleHook(r io.Reader, w io.Writer) error {
	var payload map[string]any
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return json.NewEncoder(w).Encode(HookResponse{
			Decision: "deny",
			Reason:   "invalid json: " + err.Error(),
		})
	}
	name := extractToolName(payload)
	decision, reason := Decide(name)
	return json.NewEncoder(w).Encode(HookResponse{
		Decision: decision,
		Reason:   reason,
	})
}

const hookScriptContent = `#!/usr/bin/env python3
import sys, json

try:
    data = json.load(sys.stdin)
except Exception:
    data = {}

tc = data.get("toolCall") or data.get("tool_call") or data
name = str(tc.get("name") if isinstance(tc, dict) else "")

allowed = {"view_file", "read_file", "list_dir", "grep_search", "find_by_name"}

if name in allowed:
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
