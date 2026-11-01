package evidence

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

const maxLocalDiff = 200 * 1024

// LocalDiff runs `git diff HEAD` plus untracked files, limited to owned paths,
// in dir. dir must be a git checkout. owned empty means no files (empty diff).
func LocalDiff(dir string, owned []string) (string, error) {
	if len(owned) == 0 {
		return "", nil
	}
	if err := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	var diffOutput string
	if exec.Command("git", "-C", dir, "rev-parse", "--verify", "HEAD").Run() == nil {
		out, err := runGit(dir, append([]string{"diff", "HEAD", "--"}, owned...)...)
		if err != nil {
			return "", fmt.Errorf("git diff error: %w", err)
		}
		diffOutput = out
	} else {
		unstaged, err := runGit(dir, append([]string{"diff", "--"}, owned...)...)
		if err != nil {
			return "", fmt.Errorf("git diff error: %w", err)
		}
		cached, err := runGit(dir, append([]string{"diff", "--cached", "--"}, owned...)...)
		if err != nil {
			return "", fmt.Errorf("git diff error: %w", err)
		}
		diffOutput = unstaged + cached
	}
	lsOut, err := runGit(dir, append([]string{"ls-files", "--others", "--exclude-standard", "--"}, owned...)...)
	if err != nil {
		return "", fmt.Errorf("git ls-files error: %w", err)
	}
	var files []string
	for _, file := range strings.Split(strings.TrimSpace(lsOut), "\n") {
		if file != "" {
			files = append(files, file)
		}
	}
	untrackedDiff, err := buildUntrackedDiff(dir, files)
	if err != nil {
		return "", err
	}
	out := diffOutput + untrackedDiff
	if sec := outsideSection(dir, owned); sec != "" {
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += sec
	}
	if len(out) > maxLocalDiff {
		out = out[:maxLocalDiff] + "\n[truncated at 200KB]\n"
	}
	return out, nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func pathOwned(path string, owned []string) bool {
	path = strings.TrimPrefix(path, "./")
	for _, o := range owned {
		o = strings.TrimPrefix(o, "./")
		o = strings.TrimSuffix(o, "/")
		if o == "." || o == "" || path == o || strings.HasPrefix(path, o+"/") {
			return true
		}
	}
	return false
}

func porcelainPath(line string) string {
	if len(line) < 4 {
		return ""
	}
	rest := strings.TrimSpace(line[3:])
	if i := strings.Index(rest, " -> "); i >= 0 {
		rest = strings.TrimSpace(rest[i+4:])
	}
	return strings.Trim(rest, "\"")
}

func outsideSection(dir string, owned []string) string {
	raw, err := runGit(dir, "status", "--porcelain", "-uall")
	if err != nil {
		return ""
	}
	seen := map[string]bool{}
	var b strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		p := porcelainPath(line)
		if p == "" || pathOwned(p, owned) || seen[p] {
			continue
		}
		seen[p] = true
		if b.Len() == 0 {
			b.WriteString("Files changed outside OWNED_PATHS:\n")
		}
		b.WriteString("- ")
		b.WriteString(p)
		b.WriteByte('\n')
	}
	return b.String()
}

func formatUntrackedFile(dir, file string) (string, error) {
	content, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return "", fmt.Errorf("failed to read untracked file %s: %w", file, err)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- /dev/null\n+++ b/%s\n", file))
	if bytes.IndexByte(content, 0) >= 0 {
		b.WriteString(fmt.Sprintf("binary file omitted, %d bytes\n", len(content)))
		return b.String(), nil
	}
	b.Write(content)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		b.WriteString("\n\\ No newline at end of file\n")
	}
	return b.String(), nil
}

func buildUntrackedDiff(dir string, files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}
	if len(files) == 1 {
		return formatUntrackedFile(dir, files[0])
	}
	chunks := make([]string, len(files))
	errs := make([]error, len(files))
	var wg sync.WaitGroup
	wg.Add(len(files))
	for i, file := range files {
		go func(idx int, fn string) {
			defer wg.Done()
			chunks[idx], errs[idx] = formatUntrackedFile(dir, fn)
		}(i, file)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return "", err
		}
	}
	var b strings.Builder
	for _, chunk := range chunks {
		b.WriteString(chunk)
	}
	return b.String(), nil
}

// Bundle fills worker.Evidence. It does not run git or tests itself.
func Bundle(diff, verifyOut string, verifyExit int) worker.Evidence {
	return worker.Evidence{
		Diff:         diff,
		VerifyOutput: verifyOut,
		VerifyExit:   verifyExit,
	}
}
