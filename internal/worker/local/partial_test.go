package local

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type stubTimeoutRunner struct{}

func (s *stubTimeoutRunner) Run(job worker.Job) (string, string, int, error) {
	return "", "", 1, ErrTimeout
}

func initGit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	return dir
}

func TestTimeoutPartial_ModifiedAndUntracked(t *testing.T) {
	dir := initGit(t)
	f1 := filepath.Join(dir, "tracked.txt")
	f2 := filepath.Join(dir, "untracked.txt")
	_ = os.WriteFile(f1, []byte("v1"), 0644)
	cmd := exec.Command("git", "-C", dir, "add", "tracked.txt")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-C", dir, "commit", "-m", "init")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(f1, []byte("v2"), 0644)
	_ = os.WriteFile(f2, []byte("new"), 0644)

	w := New(&stubTimeoutRunner{})
	res := w.Start(worker.Job{Repo: dir, Paths: []string{dir}})
	if res.Outcome != worker.OutcomeTimedOut {
		t.Fatalf("expected TimedOut, got %v", res.Outcome)
	}
	if !strings.Contains(res.Report, "tracked.txt") || !strings.Contains(res.Report, "untracked.txt") {
		t.Fatalf("expected report to name both files, got:\n%s", res.Report)
	}
	if !strings.Contains(res.Report, "Next round should do only the smallest remaining slice.") {
		t.Fatalf("expected next round advice, got:\n%s", res.Report)
	}
}

func TestTimeoutPartial_CleanRepo(t *testing.T) {
	dir := initGit(t)
	f1 := filepath.Join(dir, "clean.txt")
	_ = os.WriteFile(f1, []byte("v1"), 0644)
	cmd := exec.Command("git", "-C", dir, "add", "clean.txt")
	_ = cmd.Run()
	cmd = exec.Command("git", "-C", dir, "commit", "-m", "init")
	_ = cmd.Run()

	w := New(&stubTimeoutRunner{})
	res := w.Start(worker.Job{Repo: dir, Paths: []string{dir}})
	if res.Outcome != worker.OutcomeTimedOut {
		t.Fatalf("expected TimedOut, got %v", res.Outcome)
	}
	if !strings.Contains(res.Report, "no files changed") {
		t.Fatalf("expected 'no files changed', got:\n%s", res.Report)
	}
}

func TestTimeoutPartial_NotGitDir(t *testing.T) {
	dir := t.TempDir()
	w := New(&stubTimeoutRunner{})
	res := w.Start(worker.Job{Repo: dir, Paths: []string{dir}})
	if res.Outcome != worker.OutcomeTimedOut {
		t.Fatalf("expected TimedOut, got %v", res.Outcome)
	}
	if !strings.Contains(strings.ToLower(res.Report), "timed out") {
		t.Fatalf("expected report to mention timeout, got:\n%s", res.Report)
	}
}

func TestTimeoutPartial_CappedAt30(t *testing.T) {
	dir := initGit(t)
	for i := 0; i < 35; i++ {
		_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%02d.txt", i)), []byte("x"), 0644)
	}
	w := New(&stubTimeoutRunner{})
	res := w.Start(worker.Job{Repo: dir})
	if !strings.Contains(res.Report, "... 5 more") {
		t.Fatalf("expected '... 5 more', got:\n%s", res.Report)
	}
}
