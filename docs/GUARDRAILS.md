# Guardrails

Do not break these.

## Language

- Go 1.22+.
- Launcher `bin/council`. Compiled binary `bin/council-bin` from `cmd/council`.
- No `.go` source file may exceed 200 lines.
- Packages form a directed acyclic graph. Import cycles are a build failure.

## Safety

- Strip `XAI_API_KEY`, `GOOGLE_API_KEY`, `GEMINI_API_KEY` from worker and member environments.
- Council never merges to `master`. No `auto_merge`. Merging is Sahil's manual review.
- Members stay read-only sandboxed via CLI flags.
- Leave `logs/` untracked.
- Do not touch grok-agy or the Python Council checkout.

## Simplicity

- Do not invent configuration keys.
- Do not fix coordination bugs by appending prompt sentences, regex detectors, or "type go" prompts.
- Do not detect questions with `?` or phrase lists. Local asks only via a final line `COUNCIL-QUESTION: <question>`. Jules asks only via session state `AWAITING_USER_FEEDBACK`.
- One worker per `AGREED` task. No multi-worker pipelines in one paste.

## Tests

- `go test ./...` after each change.
- No network. No live Jules API.
- Assert product invariants, not prompt wording.
