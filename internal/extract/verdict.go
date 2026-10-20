package extract

import (
	"regexp"
	"strings"
)

var verdictLine = regexp.MustCompile(`(?i)^VERDICT:\s*(PASS|FAIL)$`)

const verdictCutset = "*_` \t"

// ParseVerdict returns PASS, FAIL, or "".
func ParseVerdict(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	found := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ">") {
			continue
		}
		clean := cleanVerdictLine(trimmed)
		m := verdictLine.FindStringSubmatch(clean)
		if m != nil {
			found = strings.ToUpper(m[1])
		}
	}
	return found
}

func cleanVerdictLine(s string) string {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return strings.Trim(s, verdictCutset)
	}
	before := strings.Trim(s[:idx], verdictCutset)
	after := strings.Trim(s[idx+1:], verdictCutset)
	return before + ": " + after
}
