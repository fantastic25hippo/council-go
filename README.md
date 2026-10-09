# council-go

Go rewrite of [Council](https://github.com/fantastic25hippo/council): two live member sessions, dual `AGREED`, one worker.

The Python repo stays the fallback. This repo does not import or wrap that code.

## Repository map for agents

See [How to Use council-go](docs/HOW_TO_USE_COUNCIL-GO.md), [Guardrails](docs/GUARDRAILS.md), [Worker contract](docs/WORKER_CONTRACT.md), and [Packages](docs/PACKAGES.md).

## How to start

`grok` and `agy` must be on `PATH`. Jules needs `JULES_API_KEY` in the environment or `~/.config/council/jules.env`.

```bash
go build -o bin/council-bin ./cmd/council
```

Start the meeting using `./bin/council`. It prepends `~/.grok/bin` and `~/.local/bin` to PATH, sources `~/.config/council/jules.env` if present, asks for a folder (via zenity if GUI, else via TTY), then starts the meeting in that folder.

You can also use `./bin/council-here FOLDER` to start in a specific folder (it sets `COUNCIL_START_DIR`).
Setting `COUNCIL_START_DIR` directly skips the folder picker.

The clickable launcher is `council.desktop` in this folder. Double-click it (Allow Launching if the file manager asks). A folder picker opens; Council, Grok Build, and Agy all start in the chosen folder. Do not copy the launcher to the Desktop.

Type `hi`. Paste a task, then a blank line. Dual `AGREED` plus `LAUNCH` starts one worker. SIGHUP, SIGTERM, or interrupt cancels the in-flight worker.

## Build and test

```bash
go test ./...
```

Go 1.22+. No network in tests. No live Jules API in tests.

## Guardrails

See [docs/GUARDRAILS.md](docs/GUARDRAILS.md). Short version:

- Launcher `bin/council`; compiled binary `bin/council-bin`.
- No `.go` file over 200 lines. Packages form a DAG. No import cycles.
- No pay-per-token API keys.
- Council never merges to `master`. Jules opens PRs; Sahil reviews and merges.
- The relay owns the [worker contract](docs/WORKER_CONTRACT.md). Adapters report typed outcomes. Prompts are not the protocol.
