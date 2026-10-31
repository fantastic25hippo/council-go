package meeting

import (
	"sort"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/plan"
)

func plansEqual(a, b plan.Block) bool {
	if !strings.EqualFold(normalize(a.Worker), normalize(b.Worker)) {
		return false
	}
	if normalize(a.Change) != normalize(b.Change) {
		return false
	}
	if normalize(a.Verify) != normalize(b.Verify) {
		return false
	}
	return equalNormalizedPaths(a.OwnedPaths, b.OwnedPaths)
}

func equalNormalizedPaths(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	na := make([]string, len(a))
	for i, p := range a {
		na[i] = normalize(p)
	}
	nb := make([]string, len(b))
	for i, p := range b {
		nb[i] = normalize(p)
	}
	sort.Strings(na)
	sort.Strings(nb)
	for i := range na {
		if na[i] != nb[i] {
			return false
		}
	}
	return true
}
