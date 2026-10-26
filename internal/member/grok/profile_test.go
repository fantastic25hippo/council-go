package grok

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureSandboxProfileAppendsOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := filepath.Join(home, ".grok", "sandbox.toml")
	os.MkdirAll(filepath.Dir(p), 0755)
	os.WriteFile(p, []byte("[profiles.other]\nextends = \"workspace\""), 0600)
	for i := 0; i < 2; i++ {
		if err := EnsureSandboxProfile(); err != nil {
			t.Fatal(err)
		}
	}
	b, _ := os.ReadFile(p)
	s := string(b)
	if !strings.Contains(s, "[profiles.other]") || strings.Count(s, CouncilMemberProfile) != 1 {
		t.Fatalf("bad file: %q", s)
	}
	if !strings.Contains(s, filepath.Join(home, ".local/share/grok-agy/optmem")) {
		t.Fatalf("no absolute store path: %q", s)
	}
}
