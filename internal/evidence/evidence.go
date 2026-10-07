package evidence

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

// LocalDiff runs `git diff --` plus untracked files, limited to owned paths,
// in dir. dir must be a git checkout. owned empty means no files (empty diff).
func LocalDiff(dir string, owned []string) (string, error) {
	if len(owned) == 0 {
		return "", nil
	}

	// Verify it's a git work tree
	checkCmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	if err := checkCmd.Run(); err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	args := append([]string{"-C", dir, "diff", "--"}, owned...)
	cmd := exec.Command("git", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff error: %w", err)
	}

	diffOutput := stdout.String()

	lsArgs := append([]string{"-C", dir, "ls-files", "--others", "--exclude-standard", "--"}, owned...)
	lsCmd := exec.Command("git", lsArgs...)
	var lsStdout bytes.Buffer
	lsCmd.Stdout = &lsStdout
	if err := lsCmd.Run(); err != nil {
		return "", fmt.Errorf("git ls-files error: %w", err)
	}

	untrackedFiles := strings.Split(strings.TrimSpace(lsStdout.String()), "\n")
	var untrackedDiff strings.Builder
	for _, file := range untrackedFiles {
		if file == "" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return "", fmt.Errorf("failed to read untracked file %s: %w", file, err)
		}

		untrackedDiff.WriteString(fmt.Sprintf("--- /dev/null\n+++ b/%s\n", file))
		untrackedDiff.WriteString(string(content))
		if len(content) > 0 && content[len(content)-1] != '\n' {
			untrackedDiff.WriteString("\n\\ No newline at end of file\n")
		}
	}

	return diffOutput + untrackedDiff.String(), nil
}

// Bundle fills worker.Evidence. It does not run git or tests itself.
func Bundle(diff, verifyOut string, verifyExit int) worker.Evidence {
	return worker.Evidence{
		Diff:         diff,
		VerifyOutput: verifyOut,
		VerifyExit:   verifyExit,
	}
}
