package jules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var ErrPollTimeout = errors.New("jules poll timed out")

func (l *Live) poll(ctx context.Context) error {
	if l.sess == nil {
		return errors.New("empty session id")
	}
	sleep := l.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	interval := l.Interval
	if interval <= 0 {
		interval = pollInterval
	}
	nudgeWait := l.Nudge
	if nudgeWait <= 0 {
		nudgeWait = 600 * time.Second
	}
	nowFn := l.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	deadline := nowFn().Add(l.Timeout)

	// Ensure the client has retries.
	if l.Client != nil && l.Client.API != nil {
		l.Client.API = WithRetries(l.Client.API, sleep)
	}

	lastActivity := nowFn()
	var lastAgentMsg string
	var lastState string

	for nowFn().Before(deadline) {
		if ctx != nil && ctx.Err() != nil {
			return context.Canceled
		}

		now := nowFn()

		if err := l.sess.Refresh(); err != nil {
			return err
		}
		sum, actErr := l.Client.Activities(l.sess.SID)
		if actErr != nil {
			if l.warn != nil {
				l.warn("jules activities: " + actErr.Error())
			} else {
				fmt.Fprintln(os.Stderr, "jules activities:", actErr)
			}
		}

		activityDetected := false

		if sum.LastAgent != "" {
			l.sess.SetMessage(sum.LastAgent)
			if sum.LastAgent != lastAgentMsg {
				// Head-change wake:
				// If silent for >= headChangeQuiet, treat as activity
				if now.Sub(lastActivity) >= headChangeQuiet {
					activityDetected = true
				}
				lastAgentMsg = sum.LastAgent
			}
		}

		st := strings.ToUpper(l.sess.State())
		if st != lastState {
			activityDetected = true
			lastState = st
		}

		if activityDetected {
			lastActivity = now
		}

		if !l.approved && (sum.PlanGenerated || strings.Contains(st, "PLAN")) {
			if err := l.Client.ApprovePlan(l.sess.SID); err == nil {
				l.approved = true
				lastActivity = now
			}
		}

		switch st {
		case "COMPLETED", "FAILED", "AWAITING_USER_FEEDBACK", "PAUSED", "CANCELLED":
			if st == "COMPLETED" && l.sess.PRURL() == "" {
				l.fallbackPR()
			}
			if st == "PAUSED" {
				msg := "Jules session is PAUSED; resume or answer it on jules.google.com"
				if l.warn != nil {
					l.warn(msg)
				} else {
					fmt.Fprintln(os.Stderr, msg)
				}
			}
			return nil
		}

		if st == "IN_PROGRESS" && now.Sub(lastActivity) >= nudgeWait {
			if err := l.Client.Send(l.sess.SID, nudgeText); err == nil {
				lastActivity = now
			}
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return context.Canceled
		}
	}
	return ErrPollTimeout
}
