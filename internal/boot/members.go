package boot

import (
	"fmt"
	"time"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/member"
	"github.com/fantastic25hippo/council-go/internal/member/agy"
	"github.com/fantastic25hippo/council-go/internal/member/grok"
)

func newMember(c cfg.Member, d time.Duration) (member.Member, error) {
	if c.CLI == "" {
		return nil, fmt.Errorf("member %s: empty cli", c.ID)
	}
	switch c.Driver {
	case "grok-acp":
		return grok.New(c, d), nil
	case "agy-stream":
		return agy.New(c, d), nil
	default:
		return nil, fmt.Errorf("unknown member driver %q", c.Driver)
	}
}
