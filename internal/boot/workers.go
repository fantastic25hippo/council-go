package boot

import (
	"fmt"

	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/worker"
	"github.com/fantastic25hippo/council-go/internal/worker/jules"
	"github.com/fantastic25hippo/council-go/internal/worker/local"
)

func buildWorker(c cfg.Config, cwd string, id string) (worker.Worker, error) {
	switch id {
	case "jules":
		key := jules.LoadKey()
		if key == "" {
			return nil, nil
		}
		jw := jules.NewLive(&jules.Client{API: jules.NewHTTP(key)}, c.JulesTimeout)
		jw.SetNudge(c.JulesNudge)
		return jw, nil
	case "local":
		localCfg := workerByID(c, "local")
		return local.New(&local.AgyRunner{
			CLI:     localCfg.CLI,
			Model:   localCfg.Model,
			Cwd:     cwd,
			Timeout: c.WorkerTimeout,
		}), nil
	default:
		return nil, fmt.Errorf("unknown worker %q", id)
	}
}

// BuildWorker is an exported alias for building a worker by ID.
func BuildWorker(c cfg.Config, cwd string, id string) (worker.Worker, error) {
	return buildWorker(c, cwd, id)
}

func makeWorkers(c cfg.Config, cwd string) func(string) worker.Worker {
	return func(id string) worker.Worker {
		w, err := buildWorker(c, cwd, id)
		if err != nil {
			return nil
		}
		return w
	}
}

func workerByID(c cfg.Config, id string) cfg.Worker {
	for _, w := range c.Workers {
		if w.ID == id {
			return w
		}
	}
	return cfg.Worker{ID: id, CLI: "agy"}
}
