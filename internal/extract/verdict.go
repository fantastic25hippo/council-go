package extract

import (
	"regexp"
	"strings"
)

var verdictLine = regexp.MustCompile(`(?i)^VERDICT:\s*(PASS|FAIL)\s*$`)

// ParseVerdict returns PASS, FAIL, or "".
func ParseVerdict(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	found := ""
	for _, line := range lines {
		m := verdictLine.FindStringSubmatch(strings.TrimSpace(line))
		if m != nil {
			found = strings.ToUpper(m[1])
		}
	}
	return found
}
