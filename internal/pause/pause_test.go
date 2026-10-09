package pause_test

import (
	"testing"
	"time"

	"github.com/fantastic25hippo/council-go/internal/pause"
)

func TestController_OnSIGINT(t *testing.T) {
	c := pause.New()

	// First press -> pause
	if got := c.OnSIGINT(); got != "pause" {
		t.Errorf("expected pause, got %q", got)
	}

	// Second press within 3s -> quit
	if got := c.OnSIGINT(); got != "quit" {
		t.Errorf("expected quit, got %q", got)
	}

	c = pause.New()

	// First press -> pause
	if got := c.OnSIGINT(); got != "pause" {
		t.Errorf("expected pause, got %q", got)
	}

	// Wait > 3 seconds
	time.Sleep(3100 * time.Millisecond)

	// Second press after 3s -> already paused
	if got := c.OnSIGINT(); got != "Already paused..." {
		t.Errorf("expected Already paused..., got %q", got)
	}
}
