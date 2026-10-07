package verify

import (
	"errors"
	"fmt"
	"strings"
)

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
