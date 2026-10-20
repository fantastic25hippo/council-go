package extract

import (
	"regexp"
	"strings"
)

var agreedLine = regexp.MustCompile(`(?i)^AGREED\.?$`)

var launchLine = regexp.MustCompile(`(?i)^LAUNCH(?:\s+WORKER)?:\s*(?:worker\s*=\s*)?([a-zA-Z0-9_-]+)`)

const agreedCutset = "*_` \t"

// IsAgreed is true when the last non-empty line is AGREED (optional bold/period).
func IsAgreed(text string) bool {
	last := lastNonEmpty(text)
	if last == "" {
		return false
	}
	clean := strings.Trim(last, agreedCutset)
	if strings.HasSuffix(clean, ".") {
		clean = strings.Trim(strings.TrimSuffix(clean, "."), agreedCutset) + "."
	}
	return agreedLine.MatchString(clean)
}

// ParseLaunch returns the worker id from a LAUNCH: line, or "".
func ParseLaunch(text string) string {
	for _, line := range strings.Split(text, "\n") {
		clean := strings.TrimSpace(strings.Trim(strings.TrimSpace(line), "*"))
		m := launchLine.FindStringSubmatch(clean)
		if m != nil {
			return strings.ToLower(m[1])
		}
	}
	return ""
}

func lastNonEmpty(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			return s
		}
	}
	return ""
}
