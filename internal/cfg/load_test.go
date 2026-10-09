package cfg

import (
	"os"
	"path/filepath"
	"strings"
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
	if len(c.ExtraPath) != 2 || c.ExtraPath[0] != "~/.grok/bin" || c.ExtraPath[1] != "~/.local/bin" {
		t.Fatalf("ExtraPath %+v", c.ExtraPath)
	}
}

func TestLoadExtraPathArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "council.toml")
	if err := os.WriteFile(path, []byte(`[relay]
extra_path = ["~/a", "~/b"]
`), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.ExtraPath) != 2 || c.ExtraPath[0] != "~/a" || c.ExtraPath[1] != "~/b" {
		t.Fatalf("ExtraPath %+v", c.ExtraPath)
	}
}

func TestLoadExtraPathString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "council.toml")
	if err := os.WriteFile(path, []byte(`[relay]
extra_path = "~/.grok/bin"
`), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.ExtraPath) != 1 || c.ExtraPath[0] != "~/.grok/bin" {
		t.Fatalf("ExtraPath %+v", c.ExtraPath)
	}
}

func TestLoadDirWithCouncilConfig(t *testing.T) {
	dir := t.TempDir()
	// Create another file and set COUNCIL_CONFIG
	confPath := filepath.Join(dir, "myconf.toml")
	if err := os.WriteFile(confPath, []byte(`[relay]
worker = "local"
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COUNCIL_CONFIG", confPath)

	// Create a different file in cwd which should be ignored
	if err := os.WriteFile(filepath.Join(dir, "council.toml"), []byte(`[relay]
worker = "jules"
`), 0644); err != nil {
		t.Fatal(err)
	}

	c, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Worker != "local" {
		t.Fatalf("expected local, got %s", c.Worker)
	}
}

func TestLoadDirWithCouncilConfigMissingError(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "missing.toml")
	t.Setenv("COUNCIL_CONFIG", missingPath)
	_, err := LoadDir(dir)
	if err == nil || !strings.Contains(err.Error(), "COUNCIL_CONFIG file missing") {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestLoadDirWithLocalOverlay(t *testing.T) {
	dir := t.TempDir()
	// Base file
	if err := os.WriteFile(filepath.Join(dir, "council.toml"), []byte(`[relay]
worker = "auto"
extra_path = ["~/base"]
`), 0644); err != nil {
		t.Fatal(err)
	}
	// Overlay file
	if err := os.WriteFile(filepath.Join(dir, "council.local.toml"), []byte(`[relay]
worker = "local"
`), 0644); err != nil {
		t.Fatal(err)
	}

	c, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Worker != "local" {
		t.Fatalf("expected local, got %s", c.Worker)
	}
	if len(c.ExtraPath) != 1 || c.ExtraPath[0] != "~/base" {
		t.Fatalf("expected ExtraPath base to remain, got %v", c.ExtraPath)
	}
}

func TestPrependPATH(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Set an initial path
	t.Setenv("PATH", "/usr/bin")

	c := Config{}
	c.PrependPATH([]string{dir1, dir2, filepath.Join(t.TempDir(), "missing")})

	newPath := os.Getenv("PATH")
	parts := filepath.SplitList(newPath)

	if len(parts) < 3 {
		t.Fatalf("Expected at least 3 parts, got %v", parts)
	}
	if parts[0] != dir1 {
		t.Fatalf("Expected %s at index 0, got %s", dir1, parts[0])
	}
	if parts[1] != dir2 {
		t.Fatalf("Expected %s at index 1, got %s", dir2, parts[1])
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
