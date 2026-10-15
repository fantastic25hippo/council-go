package cfg

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Member struct {
	ID          string
	DisplayName string
	Driver      string
	CLI         string
	Model       string
	Effort      string
}

type Worker struct {
	ID          string
	DisplayName string
	Driver      string
	CLI         string
	Model       string
}

type Config struct {
	Worker          string
	CheckUnpushed   bool
	DebateTurnsEach int
	FixRoundsMax    int
	MemberTimeout   time.Duration
	WorkerTimeout   time.Duration
	JulesTimeout    time.Duration
	JulesNudge      time.Duration
	ExtraPath       []string
	Members         []Member
	Workers         []Worker
}

func Defaults() Config {
	return Config{
		Worker:          "auto",
		CheckUnpushed:   true,
		DebateTurnsEach: 3,
		FixRoundsMax:    3,
		MemberTimeout:   300 * time.Second,
		WorkerTimeout:   600 * time.Second,
		JulesTimeout:    1800 * time.Second,
		JulesNudge:      600 * time.Second,
		ExtraPath:       []string{"~/.grok/bin", "~/.local/bin"},
		Members: []Member{
			{ID: "grok", DisplayName: "Grok", Driver: "grok-acp", CLI: "grok", Model: "grok-4.6", Effort: "high"},
			{ID: "gemini", DisplayName: "Gemini", Driver: "agy-stream", CLI: "agy", Model: "Gemini 3.8 Flash (High)"},
		},
		Workers: []Worker{
			{ID: "local", DisplayName: "local agy", Driver: "agy-print", CLI: "agy", Model: "Gemini 3.8 Flash (Medium)"},
			{ID: "jules", DisplayName: "Jules", Driver: "jules-remote", CLI: "jules", Model: "Jules"},
		},
	}
}

func (c Config) MemberIDs() []string {
	out := make([]string, 0, len(c.Members))
	for _, m := range c.Members {
		out = append(out, m.ID)
	}
	return out
}

func applyMember(m *Member, k, v string) {
	switch k {
	case "id":
		m.ID = v
	case "display_name":
		m.DisplayName = v
	case "driver":
		m.Driver = v
	case "cli":
		m.CLI = v
	case "model":
		m.Model = v
	case "effort":
		m.Effort = v
	}
}

func applyWorker(w *Worker, k, v string) {
	switch k {
	case "id":
		w.ID = v
	case "display_name":
		w.DisplayName = v
	case "driver":
		w.Driver = v
	case "cli":
		w.CLI = v
	case "model":
		w.Model = v
	}
}

func (c *Config) PrependPATH(dirs []string) {
	if len(dirs) == 0 {
		return
	}
	pathEnv := os.Getenv("PATH")
	var newPaths []string

	// From last to first
	for i := len(dirs) - 1; i >= 0; i-- {
		d := dirs[i]
		if strings.HasPrefix(d, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				d = home + d[1:]
			}
		}
		d = os.ExpandEnv(d)
		if d == "" {
			continue
		}

		stat, err := os.Stat(d)
		if err != nil || !stat.IsDir() {
			continue
		}

		found := false
		for _, p := range filepath.SplitList(pathEnv) {
			if p == d {
				found = true
				break
			}
		}
		for _, p := range newPaths {
			if p == d {
				found = true
				break
			}
		}
		if !found {
			newPaths = append([]string{d}, newPaths...)
		}
	}

	if len(newPaths) > 0 {
		finalPath := strings.Join(newPaths, string(os.PathListSeparator))
		if pathEnv != "" {
			finalPath = finalPath + string(os.PathListSeparator) + pathEnv
		}
		os.Setenv("PATH", finalPath)
	}
}

func (c Config) WorkerIDs() []string {
	out := make([]string, 0, len(c.Workers))
	for _, w := range c.Workers {
		out = append(out, w.ID)
	}
	return out
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			p = home + p[1:]
		}
	}
	return os.ExpandEnv(p)
}
