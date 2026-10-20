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

func TestClearPauseResetsLastSIGINT(t *testing.T) {
	c := pause.New()
	currentTime := time.Now()
	c.SetClock(func() time.Time {
		return currentTime
	})

	// First press pauses
	if got := c.OnSIGINT(); got != "pause" {
		t.Fatalf("expected pause, got %q", got)
	}

	c.ClearPause()

	// Within 3 seconds of the first press
	currentTime = currentTime.Add(1 * time.Second)

	// After ClearPause, a single SIGINT must pause, not quit
	if got := c.OnSIGINT(); got != "pause" {
		t.Fatalf("expected pause after ClearPause, got %q", got)
	}
}
