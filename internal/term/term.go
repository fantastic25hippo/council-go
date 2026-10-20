package term

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/fantastic25hippo/council-go/internal/pause"
)

var ErrPaused = errors.New("paused")

// IO is Sahil-facing print and read. Tests inject buffers.
type IO struct {
	In    io.Reader
	Out   io.Writer
	Pause *pause.Controller
	br    *bufio.Reader
	once  sync.Once
	lines chan readResult
}

type readResult struct {
	line string
	err  error
}

func (t *IO) reader() *bufio.Reader {
	if t.br == nil {
		t.br = bufio.NewReader(t.In)
	}
	return t.br
}

func (t *IO) start() {
	t.once.Do(func() {
		t.lines = make(chan readResult, 8)
		go func() {
			for {
				if t.In == nil {
					t.lines <- readResult{err: io.EOF}
					close(t.lines)
					return
				}
				line, err := t.reader().ReadString('\n')
				if err != nil && len(line) == 0 {
					t.lines <- readResult{err: err}
					close(t.lines)
					return
				}
				t.lines <- readResult{line: strings.TrimRight(line, "\r\n")}
				if err != nil {
					t.lines <- readResult{err: err}
					close(t.lines)
					return
				}
			}
		}()
	})
}

func (t *IO) Print(speaker, text string) {
	if t.Out == nil {
		return
	}
	if speaker == "" {
		fmt.Fprintf(t.Out, "%s\n", text)
		return
	}
	fmt.Fprintf(t.Out, "[%s]\n%s\n", speaker, text)
}

func (t *IO) Read(ctx context.Context, prompt string) (string, error) {
	return t.read(ctx, prompt, true)
}

// ReadPaused reads a line while already paused. It does not return ErrPaused
// for the current pause flag (so paused> can accept c /status /jules / text).
func (t *IO) ReadPaused(ctx context.Context, prompt string) (string, error) {
	return t.read(ctx, prompt, false)
}

func (t *IO) read(ctx context.Context, prompt string, interruptible bool) (string, error) {
	if interruptible && t.Pause != nil && t.Pause.IsPaused() {
		return "", ErrPaused
	}
	if t.Out != nil && prompt != "" {
		fmt.Fprint(t.Out, prompt)
	}
	if t.In == nil {
		return "", io.EOF
	}
	t.start()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case res, ok := <-t.lines:
			if !ok {
				return "", io.EOF
			}
			return res.line, res.err
		case <-ticker.C:
			if interruptible && t.Pause != nil && t.Pause.IsPaused() {
				return "", ErrPaused
			}
		}
	}
}
