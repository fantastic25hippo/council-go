package term

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// IO is Sahil-facing print and read. Tests inject buffers.
type IO struct {
	In  io.Reader
	Out io.Writer
	br  *bufio.Reader
}

func (t *IO) reader() *bufio.Reader {
	if t.br == nil {
		t.br = bufio.NewReader(t.In)
	}
	return t.br
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

func (t *IO) Read(prompt string) (string, error) {
	if t.Out != nil && prompt != "" {
		fmt.Fprint(t.Out, prompt)
	}
	if t.In == nil {
		return "", io.EOF
	}
	line, err := t.reader().ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
