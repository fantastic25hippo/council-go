package cfg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultsJules(t *testing.T) {
	c := Defaults()
	if c.JulesTimeout != 1800*time.Second {
		t.Fatalf("expected JulesTimeout 1800s, got %v", c.JulesTimeout)
	}
	if c.JulesNudge != 600*time.Second {
		t.Fatalf("expected JulesNudge 600s, got %v", c.JulesNudge)
	}
}

func TestLoadJulesTimeoutAndNudge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "council.toml")
	body := `[relay]
jules_timeout_secs = 1800
jules_nudge_secs = 600
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.JulesTimeout != 1800*time.Second {
		t.Fatalf("expected JulesTimeout 1800s, got %v", c.JulesTimeout)
	}
	if c.JulesNudge != 600*time.Second {
		t.Fatalf("expected JulesNudge 600s, got %v", c.JulesNudge)
	}
}

func TestClampJulesNudge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "council.toml")
	body := `[relay]
jules_timeout_secs = 1800
jules_nudge_secs = 1800
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.JulesTimeout != 1800*time.Second {
		t.Fatalf("expected JulesTimeout 1800s, got %v", c.JulesTimeout)
	}
	if c.JulesNudge != 600*time.Second {
		t.Fatalf("expected clamped JulesNudge 600s, got %v", c.JulesNudge)
	}

	path2 := filepath.Join(dir, "council2.toml")
	body2 := `[relay]
jules_timeout_secs = 90
jules_nudge_secs = 120
`
	if err := os.WriteFile(path2, []byte(body2), 0644); err != nil {
		t.Fatal(err)
	}
	c2, err := Load(path2)
	if err != nil {
		t.Fatal(err)
	}
	if c2.JulesTimeout != 90*time.Second {
		t.Fatalf("expected JulesTimeout 90s, got %v", c2.JulesTimeout)
	}
	if c2.JulesNudge != 30*time.Second {
		t.Fatalf("expected clamped JulesNudge 30s, got %v", c2.JulesNudge)
	}
}
