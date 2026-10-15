# Worker contract

The relay owns this contract. Adapters implement it. Meeting code never parses worker prose to decide control flow.

## Capabilities

Declared before planning. After `AGREED`, WORKER and REPO do not change.

| Field | Local | Jules |
|---|---|---|
| `CanAskAndWait` | true | true |
| `KeepsMemoryAcrossFixes` | true | true |
| `ReachesGitHub` | false | true |
| `TouchesLocalCheckout` | true | false |
| `OpensPR` | false | true |

## Outcomes

Every run ends in exactly one:

- `DONE`
- `ASKED`
- `FAILED`
- `TIMED_OUT`
- `CANCELLED`

### Local

- Last non-empty line is `COUNCIL-QUESTION: <question>` (non-empty question) → `ASKED`. Check this first.
- Exit 0 otherwise → `DONE`.
- Nonzero exit → `FAILED`.
- Timeout → `TIMED_OUT`.
- Killed by cleanup → `CANCELLED`.

A `?` anywhere else is not a question.

### Jules

- Session `COMPLETED` → `DONE`
- Session `AWAITING_USER_FEEDBACK` → `ASKED`
- Session `FAILED` → `FAILED`
- Poll timeout → `TIMED_OUT` (session may still be running; report includes session id)
- Quit / killed → `CANCELLED`

## Interface

```
Capabilities() Capabilities
Start(job Job) Run
Continue(job Job, answer string) Run
Cancel() error
```

`Run.Report` is advisory. `Run.Question` is set only on `ASKED`. `Run.Evidence` is filled by the relay later. `Run.Raw` is debug data; never parse it to choose the outcome.

## Answers vs fixes

- An **answer** is `Continue` after `ASKED`. It is never labelled a verification failure.
- A **fix** is a new `Start` with the plan plus failure reasons after `FAILED` / `TIMED_OUT`.
