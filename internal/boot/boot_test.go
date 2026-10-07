package boot

import (
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
)

func TestNewMemberDrivers(t *testing.T) {
	_, err := newMember(cfg.Member{ID: "g", Driver: "grok-acp", CLI: "grok"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newMember(cfg.Member{ID: "a", Driver: "agy-stream", CLI: "agy"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newMember(cfg.Member{ID: "x", Driver: "nope", CLI: "x"}, time.Second); err == nil {
		t.Fatal("expected unknown driver")
	}
	if _, err := newMember(cfg.Member{ID: "x", Driver: "grok-acp"}, time.Second); err == nil {
		t.Fatal("expected empty cli")
	}
}

func TestMakeWorkersLocal(t *testing.T) {
	c := cfg.Defaults()
	fn := makeWorkers(c, t.TempDir())
	w := fn("local")
	if w == nil {
		t.Fatal("local worker")
	}
	caps := w.Capabilities()
	if !caps.TouchesLocalCheckout || caps.ReachesGitHub {
		t.Fatalf("%+v", caps)
	}
}
