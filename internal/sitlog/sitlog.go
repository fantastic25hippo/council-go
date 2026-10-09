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

const prefix = "council-"
const suffix = ".log"

// Log is one sitting's file under <cwd>/logs/. A nil Log is a no-op.
type Log struct {
	Path string
	f    *os.File
}

// Open creates <cwd>/logs/council-YYYYMMDD-HHMMSS.log (UTC). Rotates
// matching files down to Keep-1 before create. Other files in logs/ stay.
func Open(cwd string, now time.Time) (*Log, error) {
	if cwd == "" {
		return nil, fmt.Errorf("empty cwd")
	}
	dir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := rotate(dir); err != nil {
		return nil, err
	}
	stamp := now.UTC().Format("20060102-150405")
	path := filepath.Join(dir, prefix+stamp+suffix)
	for n := 2; ; n++ {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return nil, err
		}
		path = filepath.Join(dir, fmt.Sprintf("%s%s-%d%s", prefix, stamp, n, suffix))
		if n > 50 {
			return nil, fmt.Errorf("log name collision")
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	l := &Log{Path: path, f: f}
	l.Line("sys", "sitting log started")
	return l, nil
}

func rotate(dir string) error {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, prefix) && strings.HasSuffix(n, suffix) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for len(names) >= Keep {
		if err := os.Remove(filepath.Join(dir, names[0])); err != nil && !os.IsNotExist(err) {
			return err
		}
		names = names[1:]
	}
	return nil
}

func (l *Log) Line(kind, text string) {
	if l == nil || l.f == nil {
		return
	}
	t := strings.ReplaceAll(Redact(text), "\n", `\n`)
	k := strings.ReplaceAll(kind, " ", "_")
	fmt.Fprintf(l.f, "%s %s %s\n", time.Now().UTC().Format(time.RFC3339), k, t)
}

func (l *Log) Close() {
	if l == nil || l.f == nil {
		return
	}
	_ = l.f.Close()
	l.f = nil
}
