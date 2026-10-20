package meeting

import (
	"context"
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/attach"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/prompt"
	"github.com/fantastic25hippo/council-go/internal/route"
	"github.com/fantastic25hippo/council-go/internal/state"
)

func (s *Session) waitTask(ctx context.Context) error {
	s.sys("Council is ready. Paste a task, then an empty line, or quit.")
	var lines []string
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := s.readSahil(ctx, "Sahil> ")
		if err != nil {
			return err
		}
		low := strings.ToLower(strings.TrimSpace(line))
		if len(lines) == 0 && (low == "quit" || low == "exit") {
			s.Council.Phase = state.PhaseAdjourn
			return nil
		}
		if strings.TrimSpace(line) == "" {
			if len(lines) == 0 {
				continue
			}
			s.Task = strings.Join(lines, "\n")
			if req, ok := attach.Parse(s.Task); ok {
				if req.Error != "" {
					s.sys(req.Error)
					s.Task = ""
					lines = nil
					continue
				}
				if req.Incomplete != "" {
					s.sys(req.Incomplete)
					s.Task = ""
					lines = nil
					continue
				}
				if req.SessionID != "" && req.PRNumber != "" {
					s.Attach = req
					s.ResolvedWorker = "jules"
					s.sys(fmt.Sprintf("Attaching to Jules session %s and PR #%s.", req.SessionID, req.PRNumber))
					s.logLine("sahil", s.Task)
					return nil
				}
			}

			override, err := plan.ParseOverride(s.Task)
			if err != nil {
				s.sys(fmt.Sprintf("override parse error: %v", err))
				s.Task = ""
				lines = nil
				continue
			}
			resolved, err := route.Resolve(override, s.Cfg.Worker, originGitHub(s.Cwd))
			if err != nil {
				s.sys(fmt.Sprintf("worker resolve error: %v", err))
				s.Task = ""
				lines = nil
				continue
			}
			s.ResolvedWorker = resolved
			s.logLine("sahil", s.Task)
			return nil
		}
		lines = append(lines, line)
	}
}

func (s *Session) discuss(ctx context.Context) error {
	s.Council.Phase = state.PhaseDiscuss
	ids := s.workerIDs()
	caps := capsFor(s.ResolvedWorker)
	pa := prompt.Proposal(s.Task, ids, s.ResolvedWorker, caps)
	pb := prompt.Proposal(s.Task, ids, s.ResolvedWorker, caps)
	ra, rb, err := sendBoth(ctx, s.A, s.B, pa, pb)
	if err != nil {
		return err
	}
	s.say(s.A, ra)
	s.say(s.B, rb)
	latestA, latestB := ra, rb
	if done, sec, err := s.tryAgree(ctx, s.A, ra, s.B); err != nil || done {
		return err
	} else if sec != "" {
		latestB = sec
	}
	if done, sec, err := s.tryAgree(ctx, s.B, rb, s.A); err != nil || done {
		return err
	} else if sec != "" {
		latestA = sec
	}
	cap := s.Cfg.DebateTurnsEach
	if cap <= 0 {
		cap = 3
	}
	for round := 0; round < cap; round++ {
		ta, err := s.A.Send(prompt.Debate(s.B.DisplayName(), latestB))
		if err != nil {
			return err
		}
		s.say(s.A, ta)
		latestA = ta
		if done, sec, err := s.tryAgree(ctx, s.A, ta, s.B); err != nil || done {
			return err
		} else if sec != "" {
			latestB = sec
			continue
		}
		tb, err := s.B.Send(prompt.Debate(s.A.DisplayName(), latestA))
		if err != nil {
			return err
		}
		s.say(s.B, tb)
		latestB = tb
		if done, sec, err := s.tryAgree(ctx, s.B, tb, s.A); err != nil || done {
			return err
		} else if sec != "" {
			latestA = sec
		}
	}
	s.sys(fmt.Sprintf("Debate cap reached. Disputed points go to Sahil."))
	da, err := s.A.Send(prompt.Disputed(cap))
	if err != nil {
		return err
	}
	db, err := s.B.Send(prompt.Disputed(cap))
	if err != nil {
		return err
	}
	s.say(s.A, da)
	s.say(s.B, db)
	s.sys("Type a ruling, or quit.")
	ruling, err := s.readSahil(ctx, "Sahil> ")
	if err != nil {
		return err
	}
	if q := strings.ToLower(strings.TrimSpace(ruling)); q == "quit" || q == "exit" {
		s.Council.Phase = state.PhaseAdjourn
		return nil
	}
	ra, rb, err = sendBoth(ctx, s.A, s.B, prompt.Ruling(ruling), prompt.Ruling(ruling))
	if err != nil {
		return err
	}
	s.say(s.A, ra)
	s.say(s.B, rb)
	if done, sec, err := s.tryAgree(ctx, s.A, ra, s.B); err != nil || done {
		return err
	} else if sec != "" {
		latestB = sec
	}
	if done, sec, err := s.tryAgree(ctx, s.B, rb, s.A); err != nil || done {
		return err
	} else if sec != "" {
		latestA = sec
	}
	s.sys("Still no dual AGREED.")
	s.Council.Phase = state.PhaseTaskDone
	return nil
}
