package plan

import (
	"errors"
	"strings"
)

// ParseOverride scans the task for WORKER-OVERRIDE: lines.
// Allowed values are "jules", "local", "auto" (case-insensitive).
func ParseOverride(task string) (string, error) {
	lines := strings.Split(task, "\n")
	var found string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)

		if !strings.HasPrefix(upper, "WORKER-OVERRIDE:") {
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		val := strings.ToLower(strings.TrimSpace(parts[1]))
		if val != "jules" && val != "local" && val != "auto" {
			continue
		}

		if found != "" && found != val {
			return "", errors.New("conflicting WORKER-OVERRIDE values")
		}
		found = val
	}

	return found, nil
}
