package meeting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/member"
)

type delayFake struct {
	*member.Fake
	d time.Duration
}

func (d *delayFake) Send(p string) (string, error) {
	time.Sleep(d.d)
	return d.Fake.Send(p)
}

func TestSendBothKeepsOtherReplyOnError(t *testing.T) {
	a := &delayFake{
		Fake: &member.Fake{Replies: []string{"hello-a"}},
		d:    50 * time.Millisecond,
	}
	b := &member.Fake{
		SendErr: errors.New("boom"),
	}

	ra, _, err := sendBoth(context.Background(), a, b, "prompt-a", "prompt-b")
	if err == nil {
		t.Fatalf("expected err != nil, got nil")
	}
	if ra != "hello-a" {
		t.Fatalf("expected ra == %q, got %q", "hello-a", ra)
	}
	if len(a.Sent) != 1 {
		t.Fatalf("expected len(a.Sent) == 1, got %d", len(a.Sent))
	}
	if len(b.Sent) != 1 {
		t.Fatalf("expected len(b.Sent) == 1, got %d", len(b.Sent))
	}
}

func TestSendBothBothOK(t *testing.T) {
	a := &member.Fake{Replies: []string{"xa"}}
	b := &member.Fake{Replies: []string{"xb"}}

	ra, rb, err := sendBoth(context.Background(), a, b, "prompt-a", "prompt-b")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if ra != "xa" || rb != "xb" {
		t.Fatalf("expected ra=%q, rb=%q; got ra=%q, rb=%q", "xa", "xb", ra, rb)
	}
}
