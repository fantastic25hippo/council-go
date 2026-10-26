package config

import "strings"

var ForbiddenKeys = []string{
	"XAI_API_KEY",
	"GOOGLE_API_KEY",
	"GEMINI_API_KEY",
}

// StripForbidden returns a copy of env (os.Environ format KEY=VALUE)
// with any entry whose key is in ForbiddenKeys removed.
// Comparison is case-sensitive and exact on the key (the part before the first =).
// Entries without = are kept unless the whole string is a forbidden key.
// Does not mutate env.
func StripForbidden(env []string) []string {
	if len(env) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(env))

	forbiddenMap := make(map[string]bool)
	for _, k := range ForbiddenKeys {
		forbiddenMap[k] = true
	}

	for _, entry := range env {
		key := entry
		if idx := strings.IndexByte(entry, '='); idx >= 0 {
			key = entry[:idx]
		}

		if !forbiddenMap[key] {
			result = append(result, entry)
		}
	}

	return result
}

// ChildEnv returns a copy of env with forbidden keys stripped, any existing GROK_AGY
// entries dropped, and GROK_AGY=1 appended.
// Does not mutate env.
func ChildEnv(env []string) []string {
	base := StripForbidden(env)
	result := make([]string, 0, len(base)+1)
	for _, entry := range base {
		key := entry
		if idx := strings.IndexByte(entry, '='); idx >= 0 {
			key = entry[:idx]
		}
		if key != "GROK_AGY" {
			result = append(result, entry)
		}
	}
	return append(result, "GROK_AGY=1")
}
