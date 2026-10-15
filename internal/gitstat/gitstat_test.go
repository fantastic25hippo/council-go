package gitstat

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStateCleanRepo(t *testing.T) {
	// Skip real fetch in tests by overriding the fetcher.
	origFetcher := Fetcher
	defer func() { Fetcher = origFetcher }()

	fetchCalled := false
	Fetcher = func(d string) error {
		fetchCalled = true
		return nil
	}

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
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "a")
	run("git", "commit", "-m", "i")
	dirty, unpushed, err := State(dir)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatal("dirty")
	}
	if !unpushed {
		t.Fatal("no upstream should be unpushed")
	}
	if !fetchCalled {
		t.Fatal("fetcher was not called")
	}
}

func TestHeadBranch(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run("git", "init", "-b", "main")
	run("git", "config", "user.email", "t@t")
	run("git", "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "a")
	run("git", "commit", "-m", "init")

	branch := HeadBranch(dir)
	if branch != "main" {
		t.Fatalf("HeadBranch = %q, want %q", branch, "main")
	}

	nonGit := t.TempDir()
	if b := HeadBranch(nonGit); b != "" {
		t.Fatalf("HeadBranch in non-git = %q, want empty", b)
	}
}
