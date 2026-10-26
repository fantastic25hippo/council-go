package agy

import (
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
	cfg              cfg.Member
	timeout          time.Duration
	hangRestartDelay time.Duration
	cwd              string
	hookDir          string
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	msgs             chan string
	mu               sync.Mutex
}

func New(c cfg.Member, timeout time.Duration) *Member {
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	return &Member{
		cfg:              c,
		timeout:          timeout,
		hangRestartDelay: 5 * time.Second,
	}
}

func (m *Member) ID() string { return m.cfg.ID }

func (m *Member) DisplayName() string {
	if m.cfg.DisplayName != "" {
		return m.cfg.DisplayName
	}
	return m.cfg.ID
}

func buildArgv(c cfg.Member, cwd string) []string {
	cli := c.CLI
	if cli == "" {
		cli = "agy"
	}
	argv := []string{
		cli, "--sandbox", "--dangerously-skip-permissions",
		"--model", c.Model,
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--add-dir", cwd,
	}
	if c.Effort != "" {
		argv = append(argv, "--effort", c.Effort)
	}
	return argv
}

func (m *Member) Start(cwd string) error {
	if cwd != "" {
		m.cwd = cwd
	}
	if m.cmd != nil || m.stdin != nil {
		_ = m.Close()
	}
	hookDir, err := os.MkdirTemp("", "council-agy-*")
	if err != nil {
		return err
	}
	cleanup := func() {
		_ = os.RemoveAll(hookDir)
		m.hookDir = ""
	}
	if err := installHook(hookDir); err != nil {
		cleanup()
		return err
	}
	m.hookDir = hookDir
	argv := buildArgv(m.cfg, m.cwd)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = m.hookDir
	cmd.Env = config.StripForbidden(os.Environ())
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	in, err := cmd.StdinPipe()
	if err != nil {
		cleanup()
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		cleanup()
		return err
	}
	if err := cmd.Start(); err != nil {
		cleanup()
		return err
	}
	m.cmd = cmd
	m.stdin = in
	m.msgs = startReader(out)
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
	if m.hookDir != "" {
		_ = os.RemoveAll(m.hookDir)
		m.hookDir = ""
	}
	return nil
}

var _ member.Member = (*Member)(nil)
