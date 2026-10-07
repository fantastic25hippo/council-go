package worker

import "strings"

// QuestionPrefix is the only local-ask marker. It must be the last
// non-empty line of the worker report. A "?" elsewhere is not a question.
const QuestionPrefix = "COUNCIL-QUESTION:"

// ParseQuestion returns the question text when the last non-empty line
// of report is "COUNCIL-QUESTION: <question>" with a non-empty question.
func ParseQuestion(report string) (question string, asked bool) {
	normalized := strings.ReplaceAll(report, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	last := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			last = s
			break
		}
	}
	if last == "" || !strings.HasPrefix(last, QuestionPrefix) {
		return "", false
	}
	q := strings.TrimSpace(strings.TrimPrefix(last, QuestionPrefix))
	if q == "" {
		return "", false
	}
	return q, true
}
