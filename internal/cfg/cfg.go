package cfg

import "time"

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

func (c Config) WorkerIDs() []string {
	out := make([]string, 0, len(c.Workers))
	for _, w := range c.Workers {
		out = append(out, w.ID)
	}
	return out
}
