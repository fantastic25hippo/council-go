package agy

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

func (m *Member) Send(prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delay := m.hangRestartDelay
	if delay <= 0 {
		delay = 5 * time.Second
	}

	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		res, timedOut, err := m.sendOnce(prompt)
		if err != nil {
			return "", err
		}
		if !timedOut {
			return res, nil
		}
		_ = m.Close()
		if attempt < maxAttempts {
			time.Sleep(delay)
			if err := m.Start(m.cwd); err != nil {
				return "", err
			}
		}
	}
	return "", fmt.Errorf("agy prompt timeout")
}

func (m *Member) sendOnce(prompt string) (string, bool, error) {
	if m.stdin == nil || m.msgs == nil {
		return "", false, fmt.Errorf("agy not started")
	}
	if err := writeUser(m.stdin, prompt); err != nil {
		return "", false, err
	}
	timer := time.NewTimer(m.timeout)
	defer timer.Stop()
	var deltas []string
	continues := 0
	for {
		select {
		case <-timer.C:
			return "", true, nil
		case line, ok := <-m.msgs:
			if !ok {
				return "", false, io.EOF
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
					if !isStreamInterrupt(err) {
						return "", false, err
					}
					if partial := joinResult(data, deltas); partial != "" {
						return partial, false, nil
					}
					if continues >= maxStreamContinues {
						return "", false, err
					}
					continues++
					if werr := writeUser(m.stdin, streamContinuePrompt); werr != nil {
						return "", false, err
					}
					timer.Reset(m.timeout)
					deltas = deltas[:0]
					continue
				}
				if t := resultText(data); t != "" {
					return t, false, nil
				}
				if joined := strings.TrimSpace(strings.Join(deltas, "")); joined != "" {
					return joined, false, nil
				}
				return "", false, nil
			}
		}
	}
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
