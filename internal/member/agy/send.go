package agy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (m *Member) Send(prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stdin == nil || m.reader == nil {
		return "", fmt.Errorf("agy not started")
	}
	msg, err := json.Marshal(map[string]any{
		"event":   "user",
		"message": map[string]string{"content": prompt},
	})
	if err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(m.stdin, "%s\n", msg); err != nil {
		return "", err
	}
	deadline := time.Now().Add(m.timeout)
	var deltas []string
	for time.Now().Before(deadline) {
		line, err := readLine(m.reader, time.Until(deadline))
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}
		event, _ := data["event"].(string)
		switch event {
		case "step_update":
			deltas = append(deltas, stepDelta(data)...)
		case "result":
			if err := resultErr(data); err != nil {
				return "", err
			}
			if t := resultText(data); t != "" {
				return t, nil
			}
			if joined := strings.TrimSpace(strings.Join(deltas, "")); joined != "" {
				return joined, nil
			}
			return "", nil
		}
	}
	return "", fmt.Errorf("agy prompt timeout")
}

func stepDelta(data map[string]any) []string {
	step, _ := data["step_update"].(map[string]any)
	if step == nil {
		return nil
	}
	if fmt.Sprint(step["step_type"]) != "agent_response" {
		return nil
	}
	if t, ok := step["text_delta"].(string); ok && t != "" {
		return []string{t}
	}
	return nil
}

func resultText(data map[string]any) string {
	res, _ := data["result"].(map[string]any)
	if res == nil {
		return ""
	}
	t, _ := res["response"].(string)
	return strings.TrimSpace(t)
}

func resultErr(data map[string]any) error {
	res, _ := data["result"].(map[string]any)
	if res == nil {
		return nil
	}
	if fmt.Sprint(res["status"]) != "ERROR" {
		return nil
	}
	msg, _ := res["error"].(string)
	if msg == "" {
		msg = "unknown agy error"
	}
	return fmt.Errorf("agy stream-json error: %s", msg)
}
