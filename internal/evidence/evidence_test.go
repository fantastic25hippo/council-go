package evidence

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func TestBundle(t *testing.T) {
	diff := "d"
	out := "out"
	exit := 1

	expected := worker.Evidence{
		Diff:         diff,
		VerifyOutput: out,
		VerifyExit:   exit,
	}

	actual := Bundle(diff, out, exit)
	if actual != expected {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}

func setupGitRepo(t *testing.T) string {
	dir := t.TempDir()

	cmd := exec.Command("git", "-C", dir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config user.name failed: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config user.email failed: %v", err)
	}

	return dir
}

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

func TestLocalDiff_ModifiedFileOutsideOwned(t *testing.T) {
	dir := setupGitRepo(t)

	err := os.Mkdir(filepath.Join(dir, "owned"), 0755)
	if err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	err = os.Mkdir(filepath.Join(dir, "unowned"), 0755)
	if err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	filePathOwned := filepath.Join(dir, "owned", "file1.txt")
	err = os.WriteFile(filePathOwned, []byte("initial1\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	filePathUnowned := filepath.Join(dir, "unowned", "file2.txt")
	err = os.WriteFile(filePathUnowned, []byte("initial2\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	cmd := exec.Command("git", "-C", dir, "add", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "commit", "-m", "initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	err = os.WriteFile(filePathOwned, []byte("modified1\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	err = os.WriteFile(filePathUnowned, []byte("modified2\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	diff, err := LocalDiff(dir, []string{"owned"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(diff, "modified1") {
		t.Errorf("expected diff to contain 'modified1', got %q", diff)
	}
	if strings.Contains(diff, "modified2") {
		t.Errorf("expected diff to not contain 'modified2', got %q", diff)
	}
}

func TestLocalDiff_DoesNotChangeGitStatus(t *testing.T) {
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

	untrackedPath := filepath.Join(dir, "owned", "untracked.txt")
	err = os.WriteFile(untrackedPath, []byte("untracked\n"), 0644)
	if err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "status", "--porcelain")
	statusBefore, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}

	_, err = LocalDiff(dir, []string{"owned"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cmd = exec.Command("git", "-C", dir, "status", "--porcelain")
	statusAfter, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}

	if string(statusBefore) != string(statusAfter) {
		t.Errorf("expected git status to remain unchanged. Before:\n%s\nAfter:\n%s", string(statusBefore), string(statusAfter))
	}
}
