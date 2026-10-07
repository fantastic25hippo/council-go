package boot

import (
	"github.com/fantastic25hippo/council-go/internal/cfg"
	"github.com/fantastic25hippo/council-go/internal/worker"
	"github.com/fantastic25hippo/council-go/internal/worker/jules"
	"github.com/fantastic25hippo/council-go/internal/worker/local"
)

func makeWorkers(c cfg.Config, cwd string) func(string) worker.Worker {
	localCfg := workerByID(c, "local")
	key := jules.LoadKey()
	return func(id string) worker.Worker {
		if id == "jules" {
			if key == "" {
				return nil
			}
			return jules.NewLive(&jules.Client{API: jules.NewHTTP(key)}, c.WorkerTimeout)
		}
		return local.New(&local.AgyRunner{
			CLI:     localCfg.CLI,
			Model:   localCfg.Model,
			Cwd:     cwd,
			Timeout: c.WorkerTimeout,
		})
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
