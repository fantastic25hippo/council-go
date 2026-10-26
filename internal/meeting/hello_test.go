package meeting

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
)

func TestGreet(t *testing.T) {
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
		t.Fatalf("unexpected error from greet: %v", err)
	}

	if len(a.Sent) != 1 || len(b.Sent) != 1 {
		t.Fatalf("expected 1 prompt sent each, got a=%d b=%d", len(a.Sent), len(b.Sent))
	}

	for _, sent := range []string{a.Sent[0], b.Sent[0]} {
		if !strings.Contains(sent, "Please briefly greet Sahil") {
			t.Fatalf("expected greet prompt in sent prompt, got:\n%s", sent)
		}
	}
}
