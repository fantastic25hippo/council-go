package meeting

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
)

type stubMember struct {
	*member.Fake
	startDelay time.Duration
	startErr   error
}

func (s *stubMember) Start(cwd string) error {
	if s.startDelay > 0 {
		time.Sleep(s.startDelay)
	}
	if s.startErr != nil {
		return s.startErr
	}
	return s.Fake.Start(cwd)
}

func TestGreet(t *testing.T) {
	t.Run("parallel start both succeed", func(t *testing.T) {
		a := &stubMember{Fake: &member.Fake{Name: "agy", Shown: "Agy", Replies: []string{"hi"}}, startDelay: 200 * time.Millisecond}
		b := &stubMember{Fake: &member.Fake{Name: "grok", Shown: "Grok", Replies: []string{"hi"}}, startDelay: 200 * time.Millisecond}
		s := &Session{
			A:       a,
			B:       b,
			Council: &state.Council{},
			IO:      &term.IO{},
		}
		start := time.Now()
		err := s.greet(context.Background())
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if elapsed >= 350*time.Millisecond {
			t.Errorf("greet took %v, expected < 350ms", elapsed)
		}
		if a.Started != 1 || b.Started != 1 {
			t.Errorf("expected both Start called once, got a=%d b=%d", a.Started, b.Started)
		}
	})

	t.Run("A.Start fails closes B if started", func(t *testing.T) {
		a := &stubMember{Fake: &member.Fake{Name: "agy", Shown: "Agy"}, startErr: errors.New("boom A")}
		b := &stubMember{Fake: &member.Fake{Name: "grok", Shown: "Grok"}, startDelay: 200 * time.Millisecond}
		s := &Session{
			A:       a,
			B:       b,
			Council: &state.Council{},
			IO:      &term.IO{},
		}
		err := s.greet(context.Background())
		if err == nil || !strings.Contains(err.Error(), "start agy: boom A") {
			t.Fatalf("expected 'start agy: boom A', got %v", err)
		}
		if b.Closed != 1 {
			t.Errorf("expected B to be closed once, got %d", b.Closed)
		}
	})

	t.Run("B.Start fails closes A if started", func(t *testing.T) {
		a := &stubMember{Fake: &member.Fake{Name: "agy", Shown: "Agy"}, startDelay: 200 * time.Millisecond}
		b := &stubMember{Fake: &member.Fake{Name: "grok", Shown: "Grok"}, startErr: errors.New("boom B")}
		s := &Session{
			A:       a,
			B:       b,
			Council: &state.Council{},
			IO:      &term.IO{},
		}
		err := s.greet(context.Background())
		if err == nil || !strings.Contains(err.Error(), "start grok: boom B") {
			t.Fatalf("expected 'start grok: boom B', got %v", err)
		}
		if a.Closed != 1 {
			t.Errorf("expected A to be closed once, got %d", a.Closed)
		}
	})
}
