package verify

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/config"
)

// CheckSyntax validates cmd syntax via `bash -n`. Nonzero exit is an error containing "VERIFY".
func CheckSyntax(cmd string) error {
	if strings.TrimSpace(cmd) == "" {
		return errors.New("VERIFY: cmd cannot be empty")
	}
	c := exec.Command("bash", "-n", "-s", "--")
	c.Stdin = strings.NewReader(cmd)
	if err := c.Run(); err != nil {
		return fmt.Errorf("VERIFY: syntax error: %w", err)
	}
	return nil
}

// Run executes Wrap(cmd, timeoutSecs) in dir.
// Combined stdout+stderr is the output.
// exit is the process exit code (127 if the binary is missing).
// err is only for Wrap/Check failures or start failures, not for a nonzero verify exit.
func Run(dir, cmd string, timeoutSecs int) (output string, exit int, err error) {
	argv, err := Wrap(cmd, timeoutSecs)
	if err != nil {
		return "", 0, err
	}

	c := exec.Command(argv[0], argv[1:]...)
	c.Dir = dir
	c.Env = config.StripForbidden(os.Environ())
	out, runErr := c.CombinedOutput()
	output = string(out)

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return output, exitErr.ExitCode(), nil
		}
		if errors.Is(runErr, exec.ErrNotFound) {
			return output, 127, nil
		}
		return output, 0, runErr
	}

	return output, 0, nil
}

// Wrap returns argv that runs cmd under GNU timeout, once.
// timeoutSecs must be > 0. cmd is passed to bash -lc unchanged.
// Result is exactly: timeout <secs> bash -lc <cmd>
func Wrap(cmd string, timeoutSecs int) ([]string, error) {
	if timeoutSecs <= 0 {
		return nil, errors.New("timeout must be > 0")
	}

	if strings.TrimSpace(cmd) == "" {
		return nil, errors.New("VERIFY: cmd cannot be empty")
	}

	return []string{
		"timeout",
		fmt.Sprintf("%d", timeoutSecs),
		"bash",
		"-lc",
		cmd,
	}, nil
}
