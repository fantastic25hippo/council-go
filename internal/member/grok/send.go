package grok

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func (m *Member) Send(prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	req := m.nextID
	m.nextID++
	if err := writeRPC(m.stdin, map[string]any{
		"jsonrpc": "2.0", "id": req, "method": "session/prompt",
		"params": map[string]any{
			"sessionId": m.sid,
			"prompt":    []any{map[string]string{"type": "text", "text": prompt}},
		},
	}); err != nil {
		return "", err
	}
	var chunks []string
	deadline := time.Now().Add(m.timeout)
	for time.Now().Before(deadline) {
		msg, err := readRPC(m.reader)
		if err != nil {
			return "", err
		}
		method, _ := msg["method"].(string)
		id := msg["id"]
		if method != "" && id != nil && fmt.Sprint(id) != fmt.Sprint(req) {
			deny(m.stdin, id)
			continue
		}
		if method == "session/update" {
			if p, ok := msg["params"].(map[string]any); ok {
				if u, ok := p["update"].(map[string]any); ok && fmt.Sprint(u["sessionUpdate"]) == "agent_message_chunk" {
					chunks = append(chunks, chunkText(u))
				}
			}
			continue
		}
		if fmt.Sprint(id) == fmt.Sprint(req) {
			if errObj, ok := msg["error"]; ok && errObj != nil {
				return "", fmt.Errorf("grok prompt: %v", errObj)
			}
			return strings.Join(chunks, ""), nil
		}
	}
	return "", fmt.Errorf("grok prompt timeout")
}

func (m *Member) call(method string, params any) (map[string]any, error) {
	id := m.nextID
	m.nextID++
	if err := writeRPC(m.stdin, map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(m.timeout)
	for time.Now().Before(deadline) {
		msg, err := readRPC(m.reader)
		if err != nil {
			return nil, err
		}
		methodName, _ := msg["method"].(string)
		mid := msg["id"]
		if methodName != "" && mid != nil && fmt.Sprint(mid) != fmt.Sprint(id) {
			deny(m.stdin, mid)
			continue
		}
		if fmt.Sprint(mid) == fmt.Sprint(id) {
			if errObj, ok := msg["error"]; ok && errObj != nil {
				return nil, fmt.Errorf("grok %s: %v", method, errObj)
			}
			res, _ := msg["result"].(map[string]any)
			if res == nil {
				res = map[string]any{}
			}
			return res, nil
		}
	}
	return nil, fmt.Errorf("grok %s timeout", method)
}

func deny(w io.Writer, id any) {
	_ = writeRPC(w, map[string]any{
		"jsonrpc": "2.0", "id": id,
		"result": map[string]any{"outcome": "cancelled", "allowed": false, "reason": "Council member is read-only."},
	})
}

func chunkText(u map[string]any) string {
	if c, ok := u["content"].(map[string]any); ok {
		if t, ok := c["text"].(string); ok {
			return t
		}
	}
	return ""
}
