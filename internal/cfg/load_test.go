package cfg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingUsesDefaults(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Worker != "auto" || c.DebateTurnsEach != 3 || len(c.Members) != 2 {
		t.Fatalf("%+v", c)
	}
}

func TestLoadRelayAndTables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "council.toml")
	body := `[relay]
worker = "local"
debate_turns_each = 2
fix_rounds_max = 1
member_timeout_secs = 10
check_unpushed = false

[[members]]
id = "grok"
display_name = "Grok"
driver = "grok-acp"
cli = "grok"

[[members]]
id = "gemini"
display_name = "Gemini"
driver = "agy-stream"
cli = "agy"

[[workers]]
id = "local"
cli = "agy"
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Worker != "local" || c.DebateTurnsEach != 2 || c.FixRoundsMax != 1 {
		t.Fatalf("relay %+v", c)
	}
	if c.CheckUnpushed || c.MemberTimeout != 10*time.Second {
		t.Fatalf("flags %+v", c)
	}
	if len(c.Members) != 2 || c.Members[0].ID != "grok" || c.Members[1].ID != "gemini" {
		t.Fatalf("members %+v", c.Members)
	}
	if len(c.Workers) != 1 || c.Workers[0].ID != "local" {
		t.Fatalf("workers %+v", c.Workers)
	}
}
