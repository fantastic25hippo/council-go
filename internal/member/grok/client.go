package grok

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func handleRequest(w io.Writer, msg map[string]any, cwd string) error {
	id := msg["id"]
	method, _ := msg["method"].(string)
	params, _ := msg["params"].(map[string]any)
	if params == nil {
		params = map[string]any{}
	}

	switch method {
	case "fs/read_text_file":
		return handleReadTextFile(w, id, params, cwd)
	case "session/request_permission":
		return handleRequestPermission(w, id, params)
	default:
		return writeRPCError(w, id, -32601, "Method not found")
	}
}

func handleReadTextFile(w io.Writer, id any, params map[string]any, cwd string) error {
	pathStr, _ := params["path"].(string)
	if pathStr == "" {
		return writeRPCError(w, id, -32602, "missing path")
	}

	cleanCwd := filepath.Clean(cwd)
	var absPath string
	if filepath.IsAbs(pathStr) {
		absPath = filepath.Clean(pathStr)
	} else {
		absPath = filepath.Clean(filepath.Join(cleanCwd, pathStr))
	}

	rel, err := filepath.Rel(cleanCwd, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return writeRPCError(w, id, -32602, "path outside cwd")
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return writeRPCError(w, id, -32603, err.Error())
	}

	lineVal, hasLine := toInt(params["line"])
	limitVal, hasLimit := toInt(params["limit"])
	var lineOpt, limitOpt *int
	if hasLine {
		lineOpt = &lineVal
	}
	if hasLimit {
		limitOpt = &limitVal
	}

	content := sliceLines(string(data), lineOpt, limitOpt)
	return writeRPC(w, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  map[string]any{"content": content},
	})
}

func sliceLines(content string, lineOpt *int, limitOpt *int) string {
	if lineOpt == nil && limitOpt == nil {
		return content
	}
	startLine := 1
	if lineOpt != nil && *lineOpt > 0 {
		startLine = *lineOpt
	}
	var lines []string
	r := bufio.NewReader(strings.NewReader(content))
	for {
		l, err := r.ReadString('\n')
		if l != "" {
			lines = append(lines, l)
		}
		if err != nil {
			break
		}
	}
	if startLine > len(lines) {
		return ""
	}
	endLine := len(lines)
	if limitOpt != nil && *limitOpt >= 0 {
		if startLine-1+*limitOpt < endLine {
			endLine = startLine - 1 + *limitOpt
		}
	}
	return strings.Join(lines[startLine-1:endLine], "")
}

func handleRequestPermission(w io.Writer, id any, params map[string]any) error {
	toolCall, _ := params["toolCall"].(map[string]any)
	kind, _ := toolCall["kind"].(string)
	kind = strings.ToLower(strings.TrimSpace(kind))

	if kind == "read" || kind == "search" || kind == "fetch" {
		options, _ := params["options"].([]any)
		var allowAlwaysID any
		var selectedID any
		for _, o := range options {
			opt, ok := o.(map[string]any)
			if !ok {
				continue
			}
			optKind := strings.ToLower(fmt.Sprint(opt["kind"]))
			optID := opt["optionId"]
			if optID == nil {
				optID = opt["id"]
			}
			if optKind == "allow_once" {
				selectedID = optID
				break
			}
			if optKind == "allow_always" && allowAlwaysID == nil {
				allowAlwaysID = optID
			}
		}
		if selectedID == nil && allowAlwaysID != nil {
			selectedID = allowAlwaysID
		}
		if selectedID != nil {
			return writeRPC(w, map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]any{
					"outcome": map[string]any{
						"outcome":  "selected",
						"optionId": selectedID,
					},
				},
			})
		}
	}

	return writeRPC(w, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"outcome": map[string]any{
				"outcome": "cancelled",
			},
		},
	})
}

func writeRPCError(w io.Writer, id any, code int, msg string) error {
	return writeRPC(w, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": msg,
		},
	})
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	}
	return 0, false
}
