package extract

import (
	"strings"
)

func ParseAnswer(text string) (string, bool) {
	last := lastNonEmpty(text)
	if strings.HasPrefix(strings.ToUpper(last), "ANSWER:") {
		return strings.TrimSpace(last[7:]), true
	}
	return "", false
}
