package sitlog

import (
	"os"
	"regexp"
	"strings"
)

var envKeys = []string{
	"JULES_API_KEY",
	"XAI_API_KEY",
	"GOOGLE_API_KEY",
	"GEMINI_API_KEY",
}

var tokenShapes = []*regexp.Regexp{
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`),
	regexp.MustCompile(`xai-[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`AIza[0-9A-Za-z_-]{30,}`),
	regexp.MustCompile(`(?i)Bearer [A-Za-z0-9._~+/=-]{20,}`),
}

// Redact replaces known env secrets and token shapes with [REDACTED].
func Redact(text string) string {
	if text == "" {
		return text
	}
	for _, k := range envKeys {
		v := os.Getenv(k)
		if v != "" {
			text = strings.ReplaceAll(text, v, "[REDACTED]")
		}
	}
	for _, re := range tokenShapes {
		text = re.ReplaceAllString(text, "[REDACTED]")
	}
	return text
}
