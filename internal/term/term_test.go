package term

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPrintAndRead(t *testing.T) {
	in := strings.NewReader("hi\n")
	var out bytes.Buffer
	io := &IO{In: in, Out: &out}
	io.Print("Grok", "hello")
	got, err := io.Read(context.Background(), "Sahil> ")
	if err != nil || got != "hi" {
		t.Fatalf("read %q %v", got, err)
	}
	s := out.String()
	if !strings.Contains(s, "[Grok]") || !strings.Contains(s, "hello") || !strings.Contains(s, "Sahil> ") {
		t.Fatalf("out %q", s)
	}
}
