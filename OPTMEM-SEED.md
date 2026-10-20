council-go (fantastic25hippo/council-go, private, Go 1.22+) is the Go rewrite of Council: two live member sessions (Grok and Gemini via agy) debate a task to dual AGREED, then exactly one worker runs it.
The legacy Python repo fantastic25hippo/council is reference only; council-go does not import or wrap it, and nothing in council-go may touch grok-agy or the Python checkout.
The default branch is master; work lands through PRs. Council itself never merges and there is no auto_merge; merging is Sahil's manual review step.
Hard rule: no .go file may exceed 200 lines (enforced by internal/check, MaxGoFileLines=200), and packages must form a DAG with no import cycles.
Build with `go build -o bin/council-bin ./cmd/council`; test with `go test ./...` after every change. Tests use no network and never call the live Jules API.
Start with `./bin/council` (folder picker) or `./bin/council-here FOLDER`; COUNCIL_START_DIR skips the picker. Jules needs JULES_API_KEY in the env or ~/.config/council/jules.env.
Config is council.toml, overridden by gitignored council.local.toml. Do not invent config keys. Members are grok (driver grok-acp, grok-4.6 high) and gemini (driver agy-stream); workers are local (agy-print) and jules (jules-remote).
Routing: a folder with a GitHub origin goes to the Jules worker, otherwise to local agy. Before Jules, Council refuses a dirty working tree or unpushed commits, because Jules only sees pushed code.
Run flow: the task goes to both members, they debate up to debate_turns_each=3, both say AGREED, one worker runs, both members PASS/FAIL the evidence; FAIL starts a fix round (fix_rounds_max=3); a split goes to Sahil.
Layout: cmd/council/main.go is the entry point; internal/meeting holds the run loop (discuss, agree, run_worker, review, fail, split, attach); internal/member/{grok,agy} are member adapters; internal/worker/{local,jules} are workers.
Other internal packages: cfg/config (settings), relay (turns and caps), extract (AGREED, plan, verdict parsing), plan, prompt, evidence, verify, gate, route, gitstat, sitlog (logs), pause, state, proc, term.
Pasting a Jules session link or ID plus a PR link or #N into the task skips the debate and attaches to that existing session and PR (internal/attach, meeting/attach.go).
Safety: XAI_API_KEY, GOOGLE_API_KEY and GEMINI_API_KEY are stripped from member and worker environments (subscriptions only); members stay read-only sandboxed via CLI flags.
Don't fix coordination bugs with extra prompt sentences, regex or '?' detectors; local workers ask only through a final COUNCIL-QUESTION: line, and Jules asks only through AWAITING_USER_FEEDBACK.
The relay owns docs/WORKER_CONTRACT.md: adapters report typed outcomes and prompts are not the protocol. Sitting logs go in an untracked logs/ folder that must be left alone.
