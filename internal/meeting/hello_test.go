package meeting

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
)

func createFakeMemo(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	binPath := filepath.Join(dir, "grok-agy-memo")
	content := "#!/bin/sh\n" + script
	if err := os.WriteFile(binPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

func TestGreetWithWake(t *testing.T) {
	script := `
if [ "$1" = "wake" ]; then
    echo "seed project memo memory"
    exit 0
fi
exit 0
`
	dir := createFakeMemo(t, script)

	a := &member.Fake{Name: "alice", Shown: "Alice", Replies: []string{"hello from alice"}}
	b := &member.Fake{Name: "bob", Shown: "Bob", Replies: []string{"hello from bob"}}
	var out bytes.Buffer
	s := &Session{
		A:       a,
		B:       b,
		Council: &state.Council{},
		IO:      &term.IO{In: strings.NewReader(""), Out: &out},
		Cwd:     dir,
	}

	if err := s.greet(context.Background()); err != nil {
		t.Fatalf("unexpected error from greet: %v", err)
	}

	if len(a.Sent) != 1 || len(b.Sent) != 1 {
		t.Fatalf("expected 1 prompt sent each, got a=%d b=%d", len(a.Sent), len(b.Sent))
	}

	for _, sent := range []string{a.Sent[0], b.Sent[0]} {
		if !strings.Contains(sent, "seed project memo memory") {
			t.Fatalf("expected wake stdout in sent prompt, got:\n%s", sent)
		}
		if !strings.Contains(sent, "Please briefly greet Sahil") {
			t.Fatalf("expected greet prompt in sent prompt, got:\n%s", sent)
		}
	}
}

func TestGreetMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	a := &member.Fake{Name: "alice", Shown: "Alice", Replies: []string{"hello from alice"}}
	b := &member.Fake{Name: "bob", Shown: "Bob", Replies: []string{"hello from bob"}}
	var out bytes.Buffer
	s := &Session{
		A:       a,
		B:       b,
		Council: &state.Council{},
		IO:      &term.IO{In: strings.NewReader(""), Out: &out},
		Cwd:     t.TempDir(),
	}

	if err := s.greet(context.Background()); err != nil {
		t.Fatalf("greet failed with missing binary: %v", err)
	}

	if len(a.Sent) != 1 || len(b.Sent) != 1 {
		t.Fatalf("expected 1 prompt sent each, got a=%d b=%d", len(a.Sent), len(b.Sent))
	}

	for _, sent := range []string{a.Sent[0], b.Sent[0]} {
		if strings.Contains(sent, "Project memory:") {
			t.Fatalf("unexpected memory block in sent prompt: %s", sent)
		}
		if !strings.Contains(sent, "Please briefly greet Sahil") {
			t.Fatalf("expected greet prompt in sent prompt, got:\n%s", sent)
		}
	}
	if !strings.Contains(out.String(), "grok-agy-memo not found") {
		t.Fatalf("expected warning in output, got: %q", out.String())
	}
}
