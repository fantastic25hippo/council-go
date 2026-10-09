package sitlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const Keep = 15

type Log struct {
	f *os.File
}

func Open(cwd string, now time.Time) (*Log, error) {
	logsDir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(logsDir)
	if err == nil {
		var matching []os.DirEntry
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "council-") && strings.HasSuffix(e.Name(), ".log") {
				matching = append(matching, e)
			}
		}

		if len(matching) >= Keep {
			sort.Slice(matching, func(i, j int) bool {
				return matching[i].Name() < matching[j].Name()
			})
			toDelete := len(matching) - (Keep - 1)
			for i := 0; i < toDelete; i++ {
				os.Remove(filepath.Join(logsDir, matching[i].Name()))
			}
		}
	}

	baseName := fmt.Sprintf("council-%s.log", now.UTC().Format("20060102-150405"))
	name := baseName
	suffix := 2
	for {
		path := filepath.Join(logsDir, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err != nil {
					return nil, err
				}
				return &Log{f: f}, nil
			}
			return nil, err
		}
		name = fmt.Sprintf("council-%s-%d.log", now.UTC().Format("20060102-150405"), suffix)
		suffix++
	}
}

func (l *Log) Line(kind, text string) {
	if l == nil || l.f == nil {
		return
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	l.f.WriteString(fmt.Sprintf("%s %s %s\n", ts, kind, Redact(text)))
}

func (l *Log) Close() {
	if l != nil && l.f != nil {
		l.f.Close()
	}
}

func Redact(text string) string {
	keys := []string{"JULES_API_KEY", "XAI_API_KEY", "GOOGLE_API_KEY", "GEMINI_API_KEY"}
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			text = strings.ReplaceAll(text, v, "[REDACTED]")
		}
	}

	// Simple heuristic redaction for shapes
	prefixes := []string{"ghp_", "gho_", "ghu_", "ghs_", "ghr_"}
	for _, p := range prefixes {
		start := 0
		for {
			idx := strings.Index(text[start:], p)
			if idx == -1 {
				break
			}
			idx += start
			end := idx + len(p)
			for end < len(text) && isAlnum(text[end]) {
				end++
			}
			if end-idx >= len(p)+20 {
				text = text[:idx] + "[REDACTED]" + text[end:]
				start = idx + len("[REDACTED]")
			} else {
				start = end
			}
		}
	}

	// GitHub PAT
	start := 0
	for {
		idx := strings.Index(text[start:], "github_pat_")
		if idx == -1 {
			break
		}
		idx += start
		end := idx + len("github_pat_")
		for end < len(text) && (isAlnum(text[end]) || text[end] == '_') {
			end++
		}
		if end-idx >= len("github_pat_")+20 {
			text = text[:idx] + "[REDACTED]" + text[end:]
			start = idx + len("[REDACTED]")
		} else {
			start = end
		}
	}

	// xAI
	start = 0
	for {
		idx := strings.Index(text[start:], "xai-")
		if idx == -1 {
			break
		}
		idx += start
		end := idx + len("xai-")
		for end < len(text) && isAlnum(text[end]) {
			end++
		}
		if end-idx >= len("xai-")+20 {
			text = text[:idx] + "[REDACTED]" + text[end:]
			start = idx + len("[REDACTED]")
		} else {
			start = end
		}
	}

	// Google / Gemini API key shape
	start = 0
	for {
		idx := strings.Index(text[start:], "AIza")
		if idx == -1 {
			break
		}
		idx += start
		end := idx + len("AIza")
		for end < len(text) && (isAlnum(text[end]) || text[end] == '_' || text[end] == '-') {
			end++
		}
		if end-idx >= len("AIza")+30 {
			text = text[:idx] + "[REDACTED]" + text[end:]
			start = idx + len("[REDACTED]")
		} else {
			start = end
		}
	}

	// Bearer tokens
	start = 0
	for {
		idx := strings.Index(text[start:], "Bearer ")
		if idx == -1 {
			break
		}
		idx += start
		end := idx + len("Bearer ")
		for end < len(text) && (isAlnum(text[end]) || text[end] == '_' || text[end] == '-' || text[end] == '.') {
			end++
		}
		if end-idx >= len("Bearer ")+20 {
			text = text[:idx] + "Bearer [REDACTED]" + text[end:]
			start = idx + len("Bearer [REDACTED]")
		} else {
			start = end
		}
	}

	return text
}

func isAlnum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
