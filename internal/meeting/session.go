package meeting

import (
	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/evidence"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/sitlog"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

// Session is one meeting. Members and workers are injected.
type Session struct {
	Cfg       cfg.Config
	A, B      member.Member
	IO        *term.IO
	Council   *state.Council
	Relay     *relay.Relay
	WorkerFor func(id string) worker.Worker
	Cwd       string
	GH        evidence.GHRunner
	Log       *sitlog.Log

	Task           string
	ResolvedWorker string
	Plan           plan.Block
	PlanText       string
}

func (s *Session) workerIDs() []string {
	if ids := s.Cfg.WorkerIDs(); len(ids) > 0 {
		return ids
	}
	return []string{"local", "jules"}
}

func (s *Session) sys(msg string) {
	if s.Log != nil {
		s.Log.Line("sys", msg)
	}
	s.IO.Print("", msg)
}

func (s *Session) say(m member.Member, text string) {
	if s.Log != nil {
		s.Log.Line("member", m.DisplayName()+": "+text)
	}
	s.IO.Print(m.DisplayName(), text)
}
