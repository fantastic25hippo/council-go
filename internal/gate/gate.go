package gate

import (
	"errors"
	"path"
	"strings"

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
// overlaps a Protected file. An owned entry overlaps a protected file when the
// protected path equals the owned path or lies under it (with "." = whole tree).
func ProtectedOverlap(owned []string) error {
	for _, o := range owned {
		oClean := strings.TrimPrefix(o, "./")
		oClean = strings.TrimPrefix(oClean, "/")
		oClean = strings.TrimSuffix(oClean, "/")
		oClean = path.Clean(oClean)

		for _, p := range Protected {
			pClean := strings.TrimPrefix(p, "./")
			pClean = strings.TrimPrefix(pClean, "/")
			pClean = path.Clean(pClean)

			if oClean == "." || oClean == "" || pClean == oClean || strings.HasPrefix(pClean, oClean+"/") {
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
