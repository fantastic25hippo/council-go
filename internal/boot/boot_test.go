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

func TestBuildWorkerExplicitLocal(t *testing.T) {
	c := cfg.Defaults()
	cwd := t.TempDir()

	// "local" must be ok
	w, err := buildWorker(c, cwd, "local")
	if err != nil {
		t.Fatalf("expected local worker ok, got: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil local worker")
	}

	// unknown worker (e.g. typo "lcoal") must return error
	if _, err := buildWorker(c, cwd, "lcoal"); err == nil {
		t.Fatal("expected error for typo 'lcoal', got nil")
	}

	// empty worker must return error
	if _, err := buildWorker(c, cwd, ""); err == nil {
		t.Fatal("expected error for empty worker id, got nil")
	}

	// makeWorkers function must return nil for unknown worker
	fn := makeWorkers(c, cwd)
	if w := fn("lcoal"); w != nil {
		t.Fatalf("expected makeWorkers('lcoal') to be nil, got: %v", w)
	}
	if w := fn("local"); w == nil {
		t.Fatal("expected makeWorkers('local') to be non-nil")
	}
}
