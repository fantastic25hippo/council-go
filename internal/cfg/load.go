package cfg

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Load reads a small council.toml subset. Missing file returns Defaults.
func Load(path string) (Config, error) {
	c := Defaults()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
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
					c.Members = nil
					sawMembers = true
				}
				c.Members = append(c.Members, Member{})
				curMember = &c.Members[len(c.Members)-1]
			case line == "[[workers]]":
				section = "workers"
				if !sawWorkers {
					c.Workers = nil
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
			applyRelay(&c, k, v)
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
	if len(c.Members) == 0 {
		c.Members = Defaults().Members
	}
	if len(c.Workers) == 0 {
		c.Workers = Defaults().Workers
	}
	return c, sc.Err()
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

func applyRelay(c *Config, k, v string) {
	switch k {
	case "worker":
		c.Worker = v
	case "check_unpushed":
		c.CheckUnpushed = v == "true"
	case "debate_turns_each":
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.DebateTurnsEach = n
		}
	case "fix_rounds_max":
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.FixRoundsMax = n
		}
	case "member_timeout_secs":
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.MemberTimeout = time.Duration(n) * time.Second
		}
	case "worker_timeout_secs":
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.WorkerTimeout = time.Duration(n) * time.Second
		}
	}
}

func applyMember(m *Member, k, v string) {
	switch k {
	case "id":
		m.ID = v
	case "display_name":
		m.DisplayName = v
	case "driver":
		m.Driver = v
	case "cli":
		m.CLI = v
	case "model":
		m.Model = v
	case "effort":
		m.Effort = v
	}
}

func applyWorker(w *Worker, k, v string) {
	switch k {
	case "id":
		w.ID = v
	case "display_name":
		w.DisplayName = v
	case "driver":
		w.Driver = v
	case "cli":
		w.CLI = v
	case "model":
		w.Model = v
	}
}
