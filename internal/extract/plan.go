package extract

import (
	"strings"

	"github.com/fantastic25hippo/council-go/internal/plan"
)

// ParsePlan reads a plan block that ends with AGREED. Worker, CHANGE, and
// VERIFY are required. The plan is not mutated after return.
func ParsePlan(text string) (plan.Block, bool) {
	if !IsAgreed(text) {
		return plan.Block{}, false
	}
	var b plan.Block
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		key, rest, ok := splitField(lines[i])
		if !ok {
			continue
		}
		switch key {
		case "WORKER":
			b.Worker = strings.ToLower(rest)
		case "REPO":
			b.Repo = rest
		case "CHANGE":
			b.Change = rest
		case "VERIFY":
			b.Verify = rest
		case "OWNED_PATHS":
			if rest != "" {
				b.OwnedPaths = append(b.OwnedPaths, rest)
			}
			i++
			for i < len(lines) {
				p := strings.TrimSpace(lines[i])
				if strings.HasPrefix(p, "- ") {
					b.OwnedPaths = append(b.OwnedPaths, strings.TrimSpace(p[2:]))
					i++
					continue
				}
				i--
				break
			}
		}
	}
	if b.Worker == "" || b.Change == "" || b.Verify == "" {
		return plan.Block{}, false
	}
	return b, true
}

func splitField(line string) (key, rest string, ok bool) {
	s := strings.TrimSpace(line)
	s = strings.Trim(s, "*")
	s = strings.TrimSpace(s)
	idx := strings.Index(s, ":")
	if idx <= 0 {
		return "", "", false
	}
	key = strings.ToUpper(strings.TrimSpace(strings.Trim(s[:idx], "*")))
	rest = strings.TrimSpace(s[idx+1:])
	rest = strings.Trim(rest, "*")
	rest = strings.TrimSpace(rest)
	switch key {
	case "WORKER", "REPO", "CHANGE", "VERIFY", "OWNED_PATHS":
		return key, rest, true
	default:
		return "", "", false
	}
}
