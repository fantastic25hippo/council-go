package jules

import (
	"bytes"
	"os/exec"
)

type execGH struct{}

func (execGH) Run(argv []string) (stdout, stderr string, exit int, err error) {
	if len(argv) == 0 {
		return "", "", 1, nil
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return out.String(), errb.String(), ee.ExitCode(), nil
		}
		return out.String(), errb.String(), 1, err
	}
	return out.String(), errb.String(), 0, nil
}
