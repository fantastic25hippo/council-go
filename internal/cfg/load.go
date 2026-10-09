package cfg

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// applyFile parses a file and modifies the given Config.
func applyFile(c *Config, path string, replaceTables bool) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	section := ""
	sawMembers, sawWorkers := false, false
	var curMember *Member
	var curWorker *Worker
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			curMember, curWorker = nil, nil
			switch {
			case line == "[relay]":
				section = "relay"
			case line == "[[members]]":
				section = "members"
				if !sawMembers {
					if replaceTables {
						c.Members = nil
					}
					sawMembers = true
				}
				c.Members = append(c.Members, Member{})
				curMember = &c.Members[len(c.Members)-1]
			case line == "[[workers]]":
				section = "workers"
				if !sawWorkers {
					if replaceTables {
						c.Workers = nil
					}
					sawWorkers = true
				}
				c.Workers = append(c.Workers, Worker{})
				curWorker = &c.Workers[len(c.Workers)-1]
			default:
				section = ""
			}
			continue
		}
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		switch section {
		case "relay":
			applyRelay(c, k, v)
		case "members":
			if curMember != nil {
				applyMember(curMember, k, v)
			}
		case "workers":
			if curWorker != nil {
				applyWorker(curWorker, k, v)
			}
		}
	}
	return sc.Err()
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			p = home + p[1:]
		}
	}
	return os.ExpandEnv(p)
}

// LoadDir loads the config from cwd, respecting COUNCIL_CONFIG and council.local.toml.
func LoadDir(cwd string) (Config, error) {
	basePath := filepath.Join(cwd, "council.toml")
	envPath := strings.TrimSpace(os.Getenv("COUNCIL_CONFIG"))
	envSet := false
	if envPath != "" {
		basePath = expandPath(envPath)
		envSet = true
	}

	c, err := Load(basePath)
	if err != nil {
		return c, err
	}
	if envSet {
		if _, statErr := os.Stat(basePath); os.IsNotExist(statErr) {
			return c, fmt.Errorf("COUNCIL_CONFIG file missing: %s", basePath)
		}
	}

	localPath := filepath.Join(cwd, "council.local.toml")
	if err := applyFile(&c, localPath, true); err != nil {
		return c, err
	}

	if len(c.Members) == 0 {
		c.Members = Defaults().Members
	}
	if len(c.Workers) == 0 {
		c.Workers = Defaults().Workers
	}

	return c, nil
}

// Load reads a small council.toml subset. Missing file returns Defaults.
func Load(path string) (Config, error) {
	c := Defaults()
	err := applyFile(&c, path, true)
	if len(c.Members) == 0 {
		c.Members = Defaults().Members
	}
	if len(c.Workers) == 0 {
		c.Workers = Defaults().Workers
	}
	return c, err
}

func splitKV(line string) (string, string, bool) {
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return "", "", false
	}
	k := strings.TrimSpace(line[:idx])
	v := strings.TrimSpace(line[idx+1:])
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	v = strings.Trim(v, `"`)
	return k, v, true
}

func parseArray(s string) []string {
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	s = strings.TrimSpace(s[1 : len(s)-1])
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var res []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = strings.Trim(p, "\"'")
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}


