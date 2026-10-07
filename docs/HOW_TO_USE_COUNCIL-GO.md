# What council-go is
council-go keeps one live Grok session and one live Gemini session, they agree a plan, then a worker types.

# How to start
From a checkout where `grok` and `agy` are on `PATH`, run `go build -o bin/council ./cmd/council` then `./bin/council`. The current directory is the working folder. Type `hi`, both members greet, then `Sahil>` waits for a task.

# Run a meeting
Paste a task at `Sahil>`, then a blank line. Independent proposals, up to 3 debate turns each, first `AGREED` writes the plan, second `AGREED` plus `LAUNCH` starts the worker, independent `VERDICT`, at most 3 fix rounds, leftover disputes go to Sahil. `quit` and `exit` adjourn. SIGHUP, SIGTERM, or interrupt cancels the in-flight worker.

# Plan block
Show fields `WORKER`, `REPO`, `OWNED_PATHS`, `CHANGE`, `VERIFY`, `AGREED`. `VERIFY` is one bare shell command. The last non-empty line of an agreeing turn is `AGREED`. The second member launches with `LAUNCH: <worker>` then `AGREED`.

# Workers
Ids `local` and `jules`. `WORKER-OVERRIDE` in the task text may be `jules`, `local`, or `auto`; distinct conflicting overrides are an error. Jules needs `REPO` as `owner/repo` and a clean pushed GitHub checkout. local gets `OWNED_PATHS` directories in this checkout. `worker = "auto"` plus a `github.com` origin resolves to jules.

# Jules
`JULES_API_KEY` from the environment or `~/.config/council/jules.env`. Council creates a session with `AUTO_CREATE_PR`, auto-approves a generated plan, and polls until `COMPLETED`, `FAILED`, or `AWAITING_USER_FEEDBACK`. Jules opens a PR and does not merge.

# Config
`council.toml` in the working directory, or built-in defaults if the file is missing. Relay keys: `worker`, `check_unpushed`, `debate_turns_each`, `fix_rounds_max`, `member_timeout_secs`, `worker_timeout_secs`. Member keys: `id`, `display_name`, `driver`, `cli`, `model`, `effort`. Worker keys: `id`, `display_name`, `driver`, `cli`, `model`.

# Hard constraints
Go 1.22+, no `.go` file over 200 lines, never use pay-per-token API keys, never merge to `master`, members stay read-only sandboxed, leave untracked `hook.log` `council-meeting.log` `council-selfreview.md` `council-jules-async.md` alone.

# Where the code lives
`cmd/council`, `internal/meeting`, `internal/member`, `internal/worker`, `internal/boot`, `council.toml`, tests next to the packages they cover.
