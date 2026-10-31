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

type pollLoopState struct {
	lastActivity time.Time
	lastAgentMsg string
	lastState    string
}

func (l *Live) logWarn(msg string) {
	if l.warn != nil {
		l.warn(msg)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

func (l *Live) updateActivity(ps *pollLoopState, now time.Time, sum actSummary, st string) {
	activity := false
	if sum.LastAgent != "" {
		l.sess.SetMessage(sum.LastAgent)
		if sum.LastAgent != ps.lastAgentMsg {
			if now.Sub(ps.lastActivity) >= headChangeQuiet {
				activity = true
			}
			ps.lastAgentMsg = sum.LastAgent
		}
	}
	if st != ps.lastState {
		activity = true
		ps.lastState = st
	}
	if activity {
		ps.lastActivity = now
	}
}

func (l *Live) checkApproval(ps *pollLoopState, now time.Time, sum actSummary, st string) {
	if !l.approved && (sum.PlanGenerated || strings.Contains(st, "PLAN")) {
		if err := l.Client.ApprovePlan(l.sess.SID); err == nil {
			l.approved = true
			ps.lastActivity = now
		}
	}
}

func (l *Live) handleState(st string) (terminal bool) {
	switch st {
	case "COMPLETED", "FAILED", "AWAITING_USER_FEEDBACK", "PAUSED", "CANCELLED":
		if st == "COMPLETED" && l.sess.PRURL() == "" {
			l.fallbackPR()
		}
		if st == "PAUSED" {
			l.logWarn("Jules session is PAUSED; resume or answer it on jules.google.com")
		}
		return true
	}
	return false
}

func (l *Live) checkNudge(ps *pollLoopState, now time.Time, st string, nudgeWait time.Duration) {
	if st == "IN_PROGRESS" && now.Sub(ps.lastActivity) >= nudgeWait {
		if err := l.Client.Send(l.sess.SID, nudgeText); err == nil {
			ps.lastActivity = now
		}
	}
}

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

	if l.Client != nil && l.Client.API != nil {
		l.Client.API = WithRetries(l.Client.API, sleep)
	}

	ps := &pollLoopState{lastActivity: nowFn()}

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
			l.logWarn("jules activities: " + actErr.Error())
		}

		st := strings.ToUpper(l.sess.State())
		l.updateActivity(ps, now, sum, st)
		l.checkApproval(ps, now, sum, st)

		if l.handleState(st) {
			return nil
		}

		l.checkNudge(ps, now, st, nudgeWait)

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return context.Canceled
		}
	}
	return ErrPollTimeout
}
