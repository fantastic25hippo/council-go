package evidence

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalDiff_EmptyOwned(t *testing.T) {
	dir := setupGitRepo(t)

	diff, err := LocalDiff(dir, []string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if diff != "" {
		t.Errorf("expected empty string, got %q", diff)
	}
}

func TestLocalDiff_TrackedModifiedFileUnderOwned(t *testing.T) {
	dir := setupGitRepo(t)

	err := os.Mkdir(filepath.Join(dir, "owned"), 0755)
	if err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	filePath := filepath.Join(dir, "owned", "file.txt")
	err = os.WriteFile(filePath, []byte("initial\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	cmd := exec.Command("git", "-C", dir, "add", "owned/file.txt")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "commit", "-m", "initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	err = os.WriteFile(filePath, []byte("modified\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	diff, err := LocalDiff(dir, []string{"owned"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(diff, "modified") {
		t.Errorf("expected diff to contain 'modified', got %q", diff)
	}
}

func TestLocalDiff_UntrackedFileUnderOwned(t *testing.T) {
	dir := setupGitRepo(t)

	err := os.Mkdir(filepath.Join(dir, "owned"), 0755)
	if err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	filePath := filepath.Join(dir, "owned", "untracked.txt")
	err = os.WriteFile(filePath, []byte("untracked content\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	diff, err := LocalDiff(dir, []string{"owned"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(diff, "--- /dev/null") || !strings.Contains(diff, "+++ b/owned/untracked.txt") || !strings.Contains(diff, "untracked content") {
		t.Errorf("expected diff to contain untracked file header and content, got %q", diff)
	}
}
