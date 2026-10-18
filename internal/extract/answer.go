package extract

import (
	"strings"
)

func ParseAnswer(text string) (string, bool) {
	last := lastNonEmpty(text)
	s := strings.TrimSpace(last)
	idx := strings.Index(s, ":")
	if idx <= 0 {
		return "", false
	}
	key := strings.ToUpper(strings.TrimSpace(strings.Trim(s[:idx], "*")))
	if key == "ANSWER" {
		rest := strings.TrimSpace(s[idx+1:])
		rest = strings.Trim(rest, "*")
		return strings.TrimSpace(rest), true
	}
	return "", false
}
