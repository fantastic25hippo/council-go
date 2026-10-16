package agy

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// agy sometimes ends a turn with a recoverable "stream was interrupted"
// error. If no text arrived yet, ask it to continue, a bounded number of times.
const (
	streamContinuePrompt = "Please continue the task you were working on."
	maxStreamContinues   = 2
)

func writeUser(w io.Writer, prompt string) error {
	msg, err := json.Marshal(map[string]any{
		"event":   "user",
		"message": map[string]string{"content": prompt},
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", msg)
	return err
}

func isStreamInterrupt(err error) bool {
	low := strings.ToLower(err.Error())
	return strings.Contains(low, "stream was interrupted") ||
		strings.Contains(low, "please continue the task")
}

func joinResult(data map[string]any, deltas []string) string {
	if t := resultText(data); t != "" {
		return t
	}
	return strings.TrimSpace(strings.Join(deltas, ""))
}
