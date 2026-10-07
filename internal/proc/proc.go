package proc

import (
	"errors"
	"os/exec"
	"syscall"
)

// Start runs argv in dir in a new process group (Setpgid: true).
// Does not wait. Returns the started *exec.Cmd.
func Start(argv []string, dir string) (*exec.Cmd, error) {
	if len(argv) == 0 {
		return nil, errors.New("empty argv")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// Kill sends SIGKILL to the process group of cmd (kill -pgid).
// If cmd or cmd.Process is nil, return error containing "process".
func Kill(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return errors.New("invalid process")
	}
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
