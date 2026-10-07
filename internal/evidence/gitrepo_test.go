package evidence

import (
	"os/exec"
	"testing"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
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
