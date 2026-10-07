package agy

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/config"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/proc"
)

type Member struct {
	cfg     cfg.Member
	timeout time.Duration
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	reader  *bufio.Reader
	mu      sync.Mutex
}

func New(c cfg.Member, timeout time.Duration) *Member {
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	return &Member{cfg: c, timeout: timeout}
}

func (m *Member) ID() string { return m.cfg.ID }

func (m *Member) DisplayName() string {
	if m.cfg.DisplayName != "" {
		return m.cfg.DisplayName
	}
	return m.cfg.ID
}

func (m *Member) Start(cwd string) error {
	cli := m.cfg.CLI
	if cli == "" {
		cli = "agy"
	}
	argv := []string{
		cli, "--sandbox", "--model", m.cfg.Model,
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--add-dir", cwd,
	}
	if m.cfg.Effort != "" {
		argv = append(argv, "--effort", m.cfg.Effort)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = cwd
	cmd.Env = config.StripForbidden(os.Environ())
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	m.cmd = cmd
	m.stdin = in
	m.reader = bufio.NewReader(out)
	return nil
}

func (m *Member) Close() error {
	if m.stdin != nil {
		_ = m.stdin.Close()
		m.stdin = nil
	}
	if m.cmd != nil {
		_ = proc.Kill(m.cmd)
		_, _ = m.cmd.Process.Wait()
		m.cmd = nil
	}
	return nil
}

var _ member.Member = (*Member)(nil)
