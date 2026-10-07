package prompt

import (
	"fmt"

	"github.com/fantastic25hippo/council-go/internal/plan"
	"github.com/fantastic25hippo/council-go/internal/worker"
)

func LocalJob(p plan.Block, caps worker.Capabilities) string {
	return fmt.Sprintf(`You are an AI assistant. You have been assigned to work on a task.
Worker: %s
Change: %s
Verify: %s

If you are blocked, end your reply with one line:
COUNCIL-QUESTION: <one question>
`, p.Worker, p.Change, p.Verify)
}

func JulesJob(p plan.Block, caps worker.Capabilities) string {
	return fmt.Sprintf(`You are an AI assistant. You have been assigned to work on a task.
Worker: %s
Change: %s
Verify: %s

If blocked or unsure, send a message describing the problem and the decision needed, then wait for an answer.
`, p.Worker, p.Change, p.Verify)
}

func VerifyReview(ev worker.Evidence, claim string) string {
	return fmt.Sprintf(`A worker claims to have completed the task.
Claim (worker's claim, unverified): %s
Diff: %s
Verify output: %s
Exit code: %d
`, claim, ev.Diff, ev.VerifyOutput, ev.VerifyExit)
}
