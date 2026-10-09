package launcher_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/check"
)

func TestCouncilStartDir(t *testing.T) {
	root, err := check.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	fakeRepo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(fakeRepo, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(root, "bin", "council"))
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(fakeRepo, "bin", "council")
	if err := os.WriteFile(dst, src, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBin := filepath.Join(fakeRepo, "bin", "council-bin")
	if err := os.WriteFile(fakeBin, []byte("#!/usr/bin/env bash\necho PWD=$(pwd)\necho ARGV=$@\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	start := filepath.Join(tmp, "start")
	if err := os.MkdirAll(start, 0755); err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(tmp, "work")
	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(dst, "from-start")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "DISPLAY=", "WAYLAND_DISPLAY=", "COUNCIL_START_DIR="+start)
	cmd.Stdin = nil
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("launcher failed: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "PWD="+start) {
		t.Errorf("expected PWD=%s, got:\n%s", start, out)
	}
	if !strings.Contains(out, "ARGV=from-start") {
		t.Errorf("expected ARGV=from-start, got:\n%s", out)
	}
}
