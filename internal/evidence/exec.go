package evidence

import (
	"errors"
	"os/exec"
)

// execGH runs commands natively.
type execGH struct{}

func (e execGH) Run(argv []string) (string, string, int, error) {
	if len(argv) == 0 {
		return "", "", 1, nil
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	out, err := cmd.Output()
	var ee *exec.ExitError
	if err != nil && errors.As(err, &ee) {
		return string(out), string(ee.Stderr), ee.ExitCode(), nil
	}
	return string(out), "", 0, err
}

func DefaultGHRunner() GHRunner {
	return execGH{}
}
