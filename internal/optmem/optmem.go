package optmem

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fantastic25hippo/council-go/internal/config"
)

const (
	wakeTimeout = 40 * time.Second
	noteTimeout = 40 * time.Second
)

var (
	notedMu   sync.Mutex
	notedDirs = make(map[string]bool)
)

func cleanEnv() []string {
	env := config.StripForbidden(os.Environ())
	deleteKeys := map[string]bool{
		"GROK_AGY":         true,
		"GROK_AGY_JOB_DIR": true,
		"MEMORY_DIR":       true,
	}
	out := make([]string, 0, len(env))
	for _, entry := range env {
		k := entry
		if idx := strings.IndexByte(entry, '='); idx >= 0 {
			k = entry[:idx]
		}
		if !deleteKeys[k] {
			out = append(out, entry)
		}
	}
	return out
}

func sanitizeLine(line string) string {
	line = strings.TrimSpace(line)
	if idx := strings.IndexAny(line, "\r\n"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	if len(line) > 280 {
		line = strings.ToValidUTF8(line[:280], "")
		line = strings.TrimSpace(line)
	}
	return line
}

// Wake executes `grok-agy-memo wake` in cwd.
// If missing or process fails, it returns ("", warning).
func Wake(cwd string) (string, string) {
	bin, err := exec.LookPath("grok-agy-memo")
	if err != nil {
		return "", "grok-agy-memo not found: " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), wakeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "wake")
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Env = cleanEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", "grok-agy-memo timed out"
		}
		return "", "grok-agy-memo: " + err.Error()
	}

	return stdout.String(), ""
}

// Note executes `grok-agy-memo note "<line>"` in cwd.
// If missing or process fails, it returns a warning string.
func Note(cwd, line string) string {
	line = sanitizeLine(line)

	bin, err := exec.LookPath("grok-agy-memo")
	if err != nil {
		return "grok-agy-memo not found: " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), noteTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "note", line)
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Env = cleanEnv()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "grok-agy-memo timed out"
		}
		return "grok-agy-memo: " + err.Error()
	}

	return ""
}

// NoteOnce calls Note at most once per process+cwd.
func NoteOnce(cwd, line string) string {
	key := cwd
	if cwd != "" {
		key = filepath.Clean(cwd)
	}

	notedMu.Lock()
	if notedDirs[key] {
		notedMu.Unlock()
		return ""
	}
	notedMu.Unlock()

	warn := Note(cwd, line)
	if warn == "" {
		notedMu.Lock()
		notedDirs[key] = true
		notedMu.Unlock()
	}
	return warn
}

func resetNoted() {
	notedMu.Lock()
	defer notedMu.Unlock()
	notedDirs = make(map[string]bool)
}
