package gate

import (
	"errors"
	"path"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

var (
	ErrUnpushed = errors.New("unpushed changes")
	ErrDirty    = errors.New("dirty working tree")
)

var Protected = []string{
	"hook.log",
	"council-meeting.log",
	"council-selfreview.md",
	"council-jules-async.md",
}

// ProtectedOverlap returns an error containing "protected" if any owned path
// equals a Protected name or is a slash-prefixed/suffixed form of one
// (e.g. "./hook.log", "dir/hook.log"). owned is plan OWNED_PATHS.
func ProtectedOverlap(owned []string) error {
	for _, o := range owned {
		base := path.Base(o)
		for _, p := range Protected {
			if base == p || o == p {
				return errors.New("protected path overlap: " + p)
			}
		}
	}
	return nil
}

// GitHubReady: if caps.ReachesGitHub is true and (dirty or unpushed), return
// an error containing "unpushed" if unpushed, else "dirty".
// If ReachesGitHub is false, always return nil.
func GitHubReady(caps worker.Capabilities, dirty, unpushed bool) error {
	if !caps.ReachesGitHub {
		return nil
	}
	if unpushed {
		return ErrUnpushed
	}
	if dirty {
		return ErrDirty
	}
	return nil
}
