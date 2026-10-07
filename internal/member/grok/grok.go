package grok

import (
	"bufio"
	"fmt"
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
	nextID  int
	sid     string
}

func New(c cfg.Member, timeout time.Duration) *Member {
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	return &Member{cfg: c, timeout: timeout, nextID: 1}
}

func (m *Member) ID() string { return m.cfg.ID }

func (m *Member) DisplayName() string {
	if m.cfg.DisplayName != "" {
		return m.cfg.DisplayName
	}
	return m.cfg.ID
}

func (m *Member) Start(cwd string) error {
	argv := []string{m.cfg.CLI, "--model", m.cfg.Model}
	if m.cfg.Effort != "" {
		argv = append(argv, "--effort", m.cfg.Effort)
	}
	argv = append(argv, "--sandbox", "read-only", "--no-plan", "agent", "stdio")
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
	if _, err := m.call("initialize", map[string]any{
		"protocolVersion": 1,
		"clientInfo":      map[string]string{"name": "council-go", "version": "0.1.0"},
		"capabilities":    map[string]any{},
	}); err != nil {
		m.Close()
		return err
	}
	res, err := m.call("session/new", map[string]any{"cwd": cwd, "mcpServers": []any{}})
	if err != nil {
		m.Close()
		return err
	}
	if id, ok := res["sessionId"].(string); ok {
		m.sid = id
	}
	if m.sid == "" {
		m.Close()
		return fmt.Errorf("grok session/new: no sessionId")
	}
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
