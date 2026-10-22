package agy

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(m.timeout)
					deltas = deltas[:0]
					continue
				}
				origText := resultText(data)
				if origText == "" {
					origText = strings.TrimSpace(strings.Join(deltas, ""))
				}
				denied := parseDeniedActions(data)
				if origText == "" && len(denied) == 0 && strings.EqualFold(resultStatus(data), "SUCCESS") {
					return "", false, fmt.Errorf("agy returned an empty response")
				}
				text := origText
				if len(denied) > 0 {
					line := fmt.Sprintf("[council: agy denied: %s]", strings.Join(denied, ", "))
					fmt.Fprintln(os.Stderr, line)
					if text != "" {
						text = text + "\n" + line
					} else {
						text = line
					}
				}
				return text, false, nil
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

func parseDeniedActions(data map[string]any) []string {
	res, _ := data["result"].(map[string]any)
	if res == nil {
		return nil
	}
	actions, ok := res["denied_actions"].([]any)
	if !ok || len(actions) == 0 {
		return nil
	}
	var names []string
	for _, a := range actions {
		switch v := a.(type) {
		case map[string]any:
			if dn, ok := v["display_name"].(string); ok && dn != "" {
				names = append(names, dn)
			} else if n, ok := v["name"].(string); ok && n != "" {
				names = append(names, n)
			}
		case string:
			if v != "" {
				names = append(names, v)
			}
		}
	}
	return names
}

func resultStatus(data map[string]any) string {
	res, _ := data["result"].(map[string]any)
	if res == nil {
		return ""
	}
	s, _ := res["status"].(string)
	return s
}
