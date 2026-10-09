package pause

import (
	"sync"
	"time"
)

type Controller struct {
	mu             sync.Mutex
	PauseRequested bool
	lastSIGINT     time.Time
}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) ClearPause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PauseRequested = false
}

func (c *Controller) IsPaused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.PauseRequested
}

func (c *Controller) OnSIGINT() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var diff time.Duration
	if !c.lastSIGINT.IsZero() {
		diff = now.Sub(c.lastSIGINT)
	} else {
		diff = 24 * time.Hour // large value for first press
	}
	c.lastSIGINT = now

	if diff <= 3*time.Second {
		return "quit"
	}

	if c.PauseRequested {
		return "Already paused..."
	}

	c.PauseRequested = true
	return "pause"
}
