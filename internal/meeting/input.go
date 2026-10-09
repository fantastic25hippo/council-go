package meeting

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
)

func (s *Session) readSahil(ctx context.Context, promptStr string) (string, error) {
	for {
		line, err := s.IO.Read(ctx, promptStr)
		if err == nil {
			return line, nil
		}
		if errors.Is(err, term.ErrPaused) {
			if err := s.pauseLoop(ctx); err != nil {
				return "", err
			}
			if s.Council.Phase == state.PhaseAdjourn {
				return "quit", nil
			}
			continue
		}
		return "", err
	}
}

func (s *Session) pauseLoop(ctx context.Context) error {
	s.sys("Paused. c resume, /status, /jules, quit, or type a message for both members.")
	for {
		line, err := s.IO.ReadPaused(ctx, "paused> ")
		if err != nil {
			return err
		}
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		low := strings.ToLower(t)
		if low == "c" {
			if s.IO.Pause != nil {
				s.IO.Pause.ClearPause()
			}
			return nil
		}
		if low == "quit" || low == "exit" {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		if low == "/status" {
			s.sys(fmt.Sprintf("Phase: %s", s.Council.Phase))
			s.sys(fmt.Sprintf("Worker: %s", s.Council.WorkerID))
			if s.Council.Handle.SessionID != "" {
				s.sys(fmt.Sprintf("SessionID: %s", s.Council.Handle.SessionID))
			}
			if s.Council.LastRun.Raw.SessionID != "" && s.Council.Handle.SessionID == "" {
				s.sys(fmt.Sprintf("SessionID: %s", s.Council.LastRun.Raw.SessionID))
			}
			if s.Council.LastRun.Raw.PRURL != "" {
				s.sys(fmt.Sprintf("PRURL: %s", s.Council.LastRun.Raw.PRURL))
			}
			continue
		}
		if low == "/jules" {
			if s.Council.LastRun.Raw.SessionID != "" || s.Council.LastRun.Raw.PRURL != "" {
				if s.Council.LastRun.Raw.SessionID != "" {
					s.sys(fmt.Sprintf("SessionID: %s", s.Council.LastRun.Raw.SessionID))
				}
				if s.Council.LastRun.Raw.PRURL != "" {
					s.sys(fmt.Sprintf("PRURL: %s", s.Council.LastRun.Raw.PRURL))
				}
			} else {
				s.sys("No Jules session.")
			}
			continue
		}

		// Any other non-empty line -> Interject
		ra, rb, err := sendBoth(ctx, s.A, s.B, prompt.Interject(t), prompt.Interject(t))
		if err != nil {
			return err
		}
		s.say(s.A, ra)
		s.say(s.B, rb)
	}
}
