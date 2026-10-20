package grok

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func runRequest(t *testing.T, msg map[string]any, cwd string) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	if err := handleRequest(&buf, msg, cwd); err != nil {
		t.Fatalf("handleRequest failed: %v", err)
	}
	var res map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &res); err != nil {
		t.Fatalf("unmarshal response %q: %v", buf.String(), err)
	}
	return res
}

func TestHandleReadTextFile(t *testing.T) {
	cwd := t.TempDir()
	filePath := filepath.Join(cwd, "sample.txt")
	body := "line1\nline2\nline3\nline4\n"
	if err := os.WriteFile(filePath, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	// Full file
	res := runRequest(t, map[string]any{
		"id": 1, "method": "fs/read_text_file",
		"params": map[string]any{"path": "sample.txt"},
	}, cwd)
	gotContent, _ := res["result"].(map[string]any)["content"].(string)
	if gotContent != body {
		t.Fatalf("got %q, want %q", gotContent, body)
	}

	// Line and limit
	res2 := runRequest(t, map[string]any{
		"id": 2, "method": "fs/read_text_file",
		"params": map[string]any{"path": filePath, "line": 2, "limit": 2},
	}, cwd)
	gotContent2, _ := res2["result"].(map[string]any)["content"].(string)
	wantContent2 := "line2\nline3\n"
	if gotContent2 != wantContent2 {
		t.Fatalf("got %q, want %q", gotContent2, wantContent2)
	}
}

func TestHandleReadTextFilePathOutsideCwd(t *testing.T) {
	cwd := t.TempDir()
	res := runRequest(t, map[string]any{
		"id": 10, "method": "fs/read_text_file",
		"params": map[string]any{"path": "../outside.txt"},
	}, cwd)
	errObj, ok := res["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", res)
	}
	if errObj["code"] != float64(-32602) {
		t.Fatalf("expected code -32602, got %v", errObj["code"])
	}
}

func TestHandleRequestPermissionReadSearchFetch(t *testing.T) {
	opts := []any{
		map[string]any{"optionId": "opt-always", "kind": "allow_always"},
		map[string]any{"optionId": "opt-once", "kind": "allow_once"},
	}
	// Selects allow_once first
	res := runRequest(t, map[string]any{
		"id": 20, "method": "session/request_permission",
		"params": map[string]any{
			"toolCall": map[string]any{"kind": "read"},
			"options":  opts,
		},
	}, "")
	outcome, _ := res["result"].(map[string]any)["outcome"].(map[string]any)
	if outcome["outcome"] != "selected" || outcome["optionId"] != "opt-once" {
		t.Fatalf("unexpected outcome: %v", outcome)
	}

	// Fallback to allow_always
	opts2 := []any{map[string]any{"optionId": "opt-always-only", "kind": "allow_always"}}
	res2 := runRequest(t, map[string]any{
		"id": 21, "method": "session/request_permission",
		"params": map[string]any{
			"toolCall": map[string]any{"kind": "search"},
			"options":  opts2,
		},
	}, "")
	outcome2, _ := res2["result"].(map[string]any)["outcome"].(map[string]any)
	if outcome2["outcome"] != "selected" || outcome2["optionId"] != "opt-always-only" {
		t.Fatalf("unexpected outcome: %v", outcome2)
	}

	// Otherwise cancelled
	res3 := runRequest(t, map[string]any{
		"id": 22, "method": "session/request_permission",
		"params": map[string]any{
			"toolCall": map[string]any{"kind": "fetch"},
			"options":  []any{},
		},
	}, "")
	outcome3, _ := res3["result"].(map[string]any)["outcome"].(map[string]any)
	if outcome3["outcome"] != "cancelled" {
		t.Fatalf("unexpected outcome: %v", outcome3)
	}
}

func TestHandleRequestPermissionWritesRefused(t *testing.T) {
	opts := []any{map[string]any{"optionId": "opt-1", "kind": "allow_once"}}
	res := runRequest(t, map[string]any{
		"id": 30, "method": "session/request_permission",
		"params": map[string]any{
			"toolCall": map[string]any{"kind": "write"},
			"options":  opts,
		},
	}, "")
	outcome, _ := res["result"].(map[string]any)["outcome"].(map[string]any)
	if outcome["outcome"] != "cancelled" {
		t.Fatalf("expected cancelled for write, got %v", outcome)
	}
}

func TestHandleOtherMethodsMethodNotFound(t *testing.T) {
	res := runRequest(t, map[string]any{
		"id": 40, "method": "fs/write_text_file",
		"params": map[string]any{"path": "x.txt", "content": "data"},
	}, "")
	errObj, ok := res["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error, got %v", res)
	}
	if errObj["code"] != float64(-32601) {
		t.Fatalf("expected code -32601, got %v", errObj["code"])
	}
}
