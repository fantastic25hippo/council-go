package grok

import (
	"fmt"
	"strings"
	"time"
)

func (m *Member) Send(prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	const maxRestarts = 2
	for attempt := 0; attempt <= maxRestarts; attempt++ {
		res, timedOut, err := m.sendOnce(prompt)
		if err != nil {
			return "", err
		}
		if !timedOut {
			return res, nil
		}
		_ = m.Close()
		if attempt < maxRestarts {
			if m.hangRestartDelay > 0 {
				time.Sleep(m.hangRestartDelay)
			}
			if err := m.Start(m.cwd); err != nil {
				return "", err
			}
			continue
		}
	}
	return "", fmt.Errorf("grok prompt timeout")
}

func (m *Member) sendOnce(prompt string) (string, bool, error) {
	req := m.nextID
	m.nextID++
	if err := writeRPC(m.stdin, map[string]any{
		"jsonrpc": "2.0", "id": req, "method": "session/prompt",
		"params": map[string]any{
			"sessionId": m.sid,
			"prompt":    []any{map[string]string{"type": "text", "text": prompt}},
		},
	}); err != nil {
		return "", false, err
	}
	var chunks []string
	timer := time.NewTimer(m.timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return "", true, nil
		case msg, ok := <-m.msgs:
			if !ok {
				return "", false, fmt.Errorf("grok prompt: reader closed")
			}
			method, _ := msg["method"].(string)
			id := msg["id"]
			if method != "" && id != nil && fmt.Sprint(id) != fmt.Sprint(req) {
				handleRequest(m.stdin, msg, m.cwd)
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
					return "", false, fmt.Errorf("grok prompt: %v", errObj)
				}
				return strings.Join(chunks, ""), false, nil
			}
		}
	}
}

func (m *Member) call(method string, params any) (map[string]any, error) {
	id := m.nextID
	m.nextID++
	if err := writeRPC(m.stdin, map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	timer := time.NewTimer(m.timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil, fmt.Errorf("grok %s timeout", method)
		case msg, ok := <-m.msgs:
			if !ok {
				return nil, fmt.Errorf("grok %s: reader closed", method)
			}
			methodName, _ := msg["method"].(string)
			mid := msg["id"]
			if methodName != "" && mid != nil && fmt.Sprint(mid) != fmt.Sprint(id) {
				handleRequest(m.stdin, msg, m.cwd)
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
	}
}

func chunkText(u map[string]any) string {
	if c, ok := u["content"].(map[string]any); ok {
		if t, ok := c["text"].(string); ok {
			return t
		}
	}
	return ""
}
