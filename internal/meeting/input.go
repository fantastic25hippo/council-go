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
	s.sys("Paused. Enter or c to resume, quit, or type a message for both members.")
	for {
		line, err := s.IO.ReadPaused(ctx, "paused> ")
		if err != nil {
			return err
		}
		t := strings.TrimSpace(line)
		low := strings.ToLower(t)
		if t == "" || low == "c" {
			if s.IO.Pause != nil {
				s.IO.Pause.ClearPause()
			}
			return nil
		}
		if low == "quit" || low == "exit" {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}

		// Any other text -> Interject with state block appended
		msgText := t + s.pauseStateBlock()
		ra, rb, err := sendBoth(ctx, s.A, s.B, prompt.Interject(msgText), prompt.Interject(msgText))
		if err != nil {
			return err
		}
		s.say(s.A, ra)
		s.say(s.B, rb)
	}
}

func (s *Session) pauseStateBlock() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n\n[State]\nPhase: %s\nWorker: %s", s.Council.Phase, s.Council.WorkerID))
	sid := s.Council.Handle.SessionID
	if sid == "" {
		sid = s.Council.LastRun.Raw.SessionID
	}
	if sid != "" {
		sb.WriteString(fmt.Sprintf("\nJules session id: %s", sid))
	}
	if pr := s.Council.LastRun.Raw.PRURL; pr != "" {
		sb.WriteString(fmt.Sprintf("\nPR URL: %s", pr))
	}
	return sb.String()
}
