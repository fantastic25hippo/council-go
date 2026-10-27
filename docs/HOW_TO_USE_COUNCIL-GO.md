# What council-go is
council-go keeps one live Grok session and one live Gemini session, they agree a plan, then a worker types.

# How to start
From a checkout where `grok` and `agy` are on `PATH`, run `go build -o bin/council-bin ./cmd/council` then `./bin/council`. The launcher prepends `~/.grok/bin` and `~/.local/bin`, sources `~/.config/council/jules.env` if present, and asks for a working folder (zenity when GUI, else TTY). `./bin/council-here FOLDER` or `COUNCIL_START_DIR` skips the picker. The clickable launcher is `council.desktop` in this folder. Double-click it. A folder picker opens. The working folder is the chosen directory. Both members start there. Type `hi`, both members greet, then `Sahil>` waits for a task.

# Startup time
When members finish starting up and exchange initial greetings, council reports `Council ready in X.Xs (start Ys, greeting Zs)`. `start Ys` is the wall-clock time spent launching member sessions in parallel, `greeting Zs` is the round-trip wall-clock time for exchanging initial greeting prompts across members, and `X.Xs` is the total startup duration.

# Run a meeting
Paste a task at `Sahil>`, then a blank line. The worker is resolved before proposals. Paste a Jules session URL/id and a GitHub PR URL or `#N` in the task to join that session and PR. Council never creates a session or a new PR on that path. A session without a PR is rejected and re-prompted. A PR without a session is a normal new task. No `/attach` command. Independent proposals, up to 3 debate turns each, first `AGREED` writes the plan, second `AGREED` plus `LAUNCH` starts the worker. Jules evidence is the PR, not the local checkout. Independent `VERDICT` on DONE. Dual PASS continues. Dual FAIL starts a fix round. A split (or a missing verdict) shows each member the other's reasons once; if they still disagree, Sahil rules. FAILED/TIMED_OUT asks members `FIX` or `ESCALATE`. At most 3 fix rounds, leftover disputes go to Sahil. Each sitting writes `<start-dir>/logs/council-<timestamp>.log` and keeps the newest 15 of those files. After both PASS the sitting waits for the next task; quit/exit adjourn. SIGHUP, SIGTERM, or a second Ctrl-C cancel the in-flight worker. Ctrl+C pauses instead of killing on the first press. When paused (`paused> ` prompt), `c` resumes, `/status` prints the phase and worker, `/jules` prints the session info, and other text goes to both members as an interject. Worker questions go to members first.

# Plan block
Show fields `WORKER`, `REPO`, `OWNED_PATHS`, `CHANGE`, `VERIFY`, `AGREED`. `VERIFY` is one bare shell command. The last non-empty line of an agreeing turn is `AGREED`. The second member launches with `LAUNCH: <worker>` then `AGREED`.

# Workers
Ids `local` and `jules`. `WORKER-OVERRIDE` in the task text may be `jules`, `local`, or `auto`; distinct conflicting overrides are an error. Jules needs `REPO` as `owner/repo` and a clean pushed GitHub checkout. local gets `OWNED_PATHS` directories in this checkout. `worker = "auto"` plus a `github.com` origin resolves to jules.

Local jobs run through `grok-agy-run`, enforcing the tool policy hook, tool and step limits, 96K context compaction, and a 5-job queue, so callers should bundle related work into single jobs.

# Jules
`JULES_API_KEY` from the environment or `~/.config/council/jules.env`. Council creates a session with `AUTO_CREATE_PR`, auto-approves a generated plan, and polls until `COMPLETED`, `FAILED`, or `AWAITING_USER_FEEDBACK`. Jules opens a PR and does not merge.

# Config
`council.toml` in the working directory, or built-in defaults if the file is missing. `COUNCIL_CONFIG` acts as the base file if set. A `council.local.toml` overlay in the start directory replaces keys/tables. Relay keys: `extra_path` (list of dirs prepended to PATH), `worker`, `check_unpushed`, `debate_turns_each`, `fix_rounds_max`, `member_timeout_secs`, `worker_timeout_secs`, `jules_timeout_secs`, `jules_nudge_secs`. Member keys: `id`, `display_name`, `driver`, `cli`, `model`, `effort`. Worker keys: `id`, `display_name`, `driver`, `cli`, `model`.

# Hard constraints
Go 1.22+, no `.go` file over 200 lines, never use pay-per-token API keys, never merge to `master`, members stay read-only sandboxed, leave untracked `hook.log` `council-meeting.log` `council-selfreview.md` `council-jules-async.md` and `logs/` alone.

# Where the code lives
`cmd/council`, `internal/meeting`, `internal/member`, `internal/worker`, `internal/boot`, `council.toml`, tests next to the packages they cover.
