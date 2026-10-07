package evidence

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
