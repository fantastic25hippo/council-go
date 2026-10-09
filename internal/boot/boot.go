package boot

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/meeting"
	"github.com/fantastic25hippo/council-go/internal/pause"
	"github.com/fantastic25hippo/council-go/internal/relay"
	"github.com/fantastic25hippo/council-go/internal/state"
	"github.com/fantastic25hippo/council-go/internal/term"
	"github.com/fantastic25hippo/council-go/internal/worker/local"
)

// Run loads council.toml (or defaults) and starts the meeting.
func Run(ctx context.Context, in io.Reader, out io.Writer, cwd string, pc *pause.Controller) error {
	c, err := cfg.Load(filepath.Join(cwd, "council.toml"))
	if err != nil {
		return err
	}
	if len(c.Members) < 2 {
		return fmt.Errorf("need two members")
	}
	a, err := newMember(c.Members[0], c.MemberTimeout)
	if err != nil {
		return err
	}
	b, err := newMember(c.Members[1], c.MemberTimeout)
	if err != nil {
		return err
	}
	council := &state.Council{}
	s := &meeting.Session{
		Cfg:       c,
		A:         a,
		B:         b,
		IO:        &term.IO{In: in, Out: out, Pause: pc},
		Council:   council,
		Relay:     &relay.Relay{Council: council, Worker: local.New(nil)},
		WorkerFor: makeWorkers(c, cwd),
		Cwd:       cwd,
	}
	return s.Run(ctx)
}
