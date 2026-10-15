package cfg

import (
	"strconv"
	"time"
)

func posInt(v string) int {
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return n
	}
	return 0
}

func ensureTables(c *Config) {
	if len(c.Members) == 0 {
		c.Members = Defaults().Members
	}
	if len(c.Workers) == 0 {
		c.Workers = Defaults().Workers
	}
}

func applyRelay(c *Config, k, v string) {
	switch k {
	case "worker":
		c.Worker = v
	case "extra_path":
		c.ExtraPath = parseArray(v)
	case "check_unpushed":
		c.CheckUnpushed = v == "true"
	case "debate_turns_each":
		if n := posInt(v); n > 0 {
			c.DebateTurnsEach = n
		}
	case "fix_rounds_max":
		if n := posInt(v); n > 0 {
			c.FixRoundsMax = n
		}
	case "member_timeout_secs":
		if n := posInt(v); n > 0 {
			c.MemberTimeout = time.Duration(n) * time.Second
		}
	case "worker_timeout_secs":
		if n := posInt(v); n > 0 {
			c.WorkerTimeout = time.Duration(n) * time.Second
		}
	case "jules_timeout_secs":
		if n := posInt(v); n > 0 {
			c.JulesTimeout = time.Duration(n) * time.Second
		}
	case "jules_nudge_secs":
		if n := posInt(v); n > 0 {
			c.JulesNudge = time.Duration(n) * time.Second
		}
	}
}

func clampJules(c *Config) {
	if c.JulesTimeout > 0 && c.JulesNudge >= c.JulesTimeout {
		c.JulesNudge = c.JulesTimeout / 3
	}
}
