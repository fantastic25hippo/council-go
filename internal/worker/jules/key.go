package jules

import (
	"os"
	"path/filepath"
	"strings"
)

// LoadKey reads JULES_API_KEY from the environment or ~/.config/council/jules.env.
// The value is never logged.
func LoadKey() string {
	if k := strings.TrimSpace(os.Getenv("JULES_API_KEY")); k != "" {
		return k
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(home, ".config", "council", "jules.env"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "JULES_API_KEY=") {
			continue
		}
		v := strings.TrimSpace(strings.TrimPrefix(line, "JULES_API_KEY="))
		return strings.Trim(v, `"'`)
	}
	return ""
}
