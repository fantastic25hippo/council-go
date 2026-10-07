package jules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKeyFromEnv(t *testing.T) {
	t.Setenv("JULES_API_KEY", "from-env")
	if LoadKey() != "from-env" {
		t.Fatal("env")
	}
}

func TestLoadKeyFromFile(t *testing.T) {
	t.Setenv("JULES_API_KEY", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "council")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "jules.env"), []byte("JULES_API_KEY=from-file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if LoadKey() != "from-file" {
		t.Fatal("file")
	}
}

func TestLoadKeyMissing(t *testing.T) {
	t.Setenv("JULES_API_KEY", "")
	t.Setenv("HOME", t.TempDir())
	if LoadKey() != "" {
		t.Fatal("empty")
	}
}
