package gitstat

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestState_UntrackedFilesNotDirty(t *testing.T) {
	origFetcher := Fetcher
	defer func() { Fetcher = origFetcher }()
	Fetcher = func(d string) error { return nil }

	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run("git", "init")
	run("git", "config", "user.email", "t@t")
	run("git", "config", "user.name", "t")

	tracked := filepath.Join(dir, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "tracked.txt")
	run("git", "commit", "-m", "init")

	// Untracked file added -> should NOT be dirty
	untracked := filepath.Join(dir, "untracked.log")
	if err := os.WriteFile(untracked, []byte("log data\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dirty, _, err := State(dir)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if dirty {
		t.Errorf("expected dirty=false with untracked file, got true")
	}

	// Modify tracked file -> should be dirty
	if err := os.WriteFile(tracked, []byte("hello modified\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dirty, _, err = State(dir)
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if !dirty {
		t.Errorf("expected dirty=true with modified tracked file, got false")
	}
}
