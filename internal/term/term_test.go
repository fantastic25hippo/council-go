package term

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestPrintAndRead(t *testing.T) {
	in := strings.NewReader("hi\n")
	var out bytes.Buffer
	tio := &IO{In: in, Out: &out}
	tio.Print("Grok", "hello")
	got, err := tio.Read(context.Background(), "Sahil> ")
	if err != nil || got != "hi" {
		t.Fatalf("read %q %v", got, err)
	}
	s := out.String()
	if !strings.Contains(s, "[Grok]") || !strings.Contains(s, "hello") || !strings.Contains(s, "Sahil> ") {
		t.Fatalf("out %q", s)
	}
}

func TestReadNoNewlineEOF(t *testing.T) {
	in := strings.NewReader("x")
	tio := &IO{In: in}

	got, err := tio.Read(context.Background(), "")
	if err != nil {
		t.Fatalf("first Read error: %v", err)
	}
	if got != "x" {
		t.Fatalf("first Read expected 'x', got %q", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	got2, err2 := tio.Read(ctx, "")
	if err2 != io.EOF {
		t.Fatalf("second Read expected io.EOF within 1s, got: val=%q err=%v", got2, err2)
	}
}
