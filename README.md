# council-go

Go rewrite of [Council](https://github.com/fantastic25hippo/council): a small relay that keeps two live member sessions, carries proposals, spots `AGREED`, and hands one job to a worker.

The Python repo at `/home/sahil/Desktop/council` stays the working fallback. This repo is a clean rewrite. It does not import or wrap that code.

## Guardrails

See [docs/GUARDRAILS.md](docs/GUARDRAILS.md). Short version:

- Go 1.22+. One static binary later, `bin/council`.
- No `.go` file over 200 lines. Packages form a DAG. No import cycles.
- No pay-per-token API keys.
- Council never merges to `master`. Jules opens PRs; Sahil reviews and merges.
- The relay owns the [worker contract](docs/WORKER_CONTRACT.md). Adapters report typed outcomes. Prompts are not the protocol.

## Packages

See [docs/PACKAGES.md](docs/PACKAGES.md) for a summary of the internal structure.

## Build and test

Binary:
```bash
go build -o bin/council ./cmd/council
go test ./...
```

`cmd/council` currently waits for SIGHUP/TERM/interrupt then cancels the worker. It is not the full meeting loop.

No network in tests. No live Jules API in tests.
